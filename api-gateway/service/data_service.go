package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/util"
	"stock/common/errgroup"
	"stock/common/log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/axgle/mahonia"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

// DataService 数据服务
type DataService struct {
	data *model.Data
}

var (
	dataService *DataService
	dataOnce    sync.Once
)

// DataServiceInstance 数据服务:负责各种数据抓取工作
func DataServiceInstance() *DataService {
	dataOnce.Do(func() {
		dataService = &DataService{
			data: &model.Data{},
		}
		ctx := context.Background()
		//if err := dataService.load(ctx); err != nil {
		//	panic(err)
		//}
		go func() {
			for range time.Tick(3 * time.Second) {
				now := time.Now()
				if now.Hour() < 9 || (now.Hour() == 9 && now.Minute() <= 30) {
					continue
				}
				if err := dataService.load(ctx); err != nil {
					log.Errorf("home load err:%+v", err)
					continue
				}
			}
		}()
	})
	return dataService
}

const (
	dataServiceCacheKey        = "stock_data_cache_key"
	industrySector      string = "https://82.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=200&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:90+t:2+f:!50&fields=f2,f3,f4,f8,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f26,f22,f33,f11,f62,f128,f136,f115,f152,f124,f107,f104,f105,f140,f141,f207,f208,f209,f222,f43"                // 行业版块
	conceptSector       string = "https://86.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=200&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:90+t:3+f:!50&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f26,f22,f33,f11,f62,f128,f136,f115,f152,f124,f107,f104,f105,f140,f141,f207,f208,f209,f222" // 概念版块
	upRankURL           string = "https://67.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=200&po=1&np=1&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f22,f11,f62,f128,f136,f115,f152"
	downRankURL         string = "https://67.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=200&po=0&np=1&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f22,f11,f62,f128,f136,f115,f152"
	turnoverRankURL     string = "https://67.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=200&po=1&np=1&fltt=2&invt=2&fid=f8&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f22,f11,f62,f128,f136,f115,f152"
	volumeRankURL       string = "https://67.push2.eastmoney.com/api/qt/clist/get?pn=1&pz=200&po=1&np=1&fltt=2&invt=2&fid=f6&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048&fields=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f12,f13,f14,f15,f16,f17,f18,f20,f21,f23,f24,f25,f22,f11,f62,f128,f136,f115,f152"
	netInBalanceRankURL string = "https://push2.eastmoney.com/api/qt/clist/get?fid=f62&po=1&pz=200&pn=2&np=1&fltt=2&invt=2&ut=b2884a393a59ad64002292a3e90d46a5&fs=m%3A0%2Bt%3A6%2Bf%3A!2%2Cm%3A0%2Bt%3A13%2Bf%3A!2%2Cm%3A0%2Bt%3A80%2Bf%3A!2%2Cm%3A1%2Bt%3A2%2Bf%3A!2%2Cm%3A1%2Bt%3A23%2Bf%3A!2%2Cm%3A0%2Bt%3A7%2Bf%3A!2%2Cm%3A1%2Bt%3A3%2Bf%3A!2&fields=f12%2Cf14%2Cf2%2Cf3%2Cf62%2Cf184%2Cf66%2Cf69%2Cf72%2Cf75%2Cf78%2Cf81%2Cf84%2Cf87%2Cf204%2Cf205%2Cf124%2Cf1%2Cf13"
	overViewURL         string = "https://push2ex.eastmoney.com/getTopicZDFenBu?cb=callbackdata9443058&ut=7eea3edcaed734bea9cbfc24409ed989&dpt=wz.ztzt"
)

func (s *DataService) GetData() *model.Data {
	if time.Since(s.data.UpdateTime).Seconds() > 10 {
		// 从缓存中读取数据
		buff, err := db.Get(context.Background(), dataServiceCacheKey).Bytes()
		if err != nil {
			return &model.Data{}
		}
		result := &model.Data{}
		err = json.Unmarshal(buff, result)
		if err != nil {
			log.Errorf("json.Unmarshal error: %v", err)
			return &model.Data{}
		}
		return result
	}
	return s.data
}

// 数据加载
func (s *DataService) load(ctx context.Context) error {
	var mutex sync.Mutex
	wg := errgroup.GroupWithCount(8)
	// 市场概况数据
	wg.Go(func() error {
		ret, err := s.getOverview()
		if err != nil {
			return err
		}
		mutex.Lock()
		s.data.MarketOverView = ret
		mutex.Unlock()
		return nil
	})

	// 股票指标排名排名数据
	wg.Go(func() error {
		ret, err := s.stockIndicatorRank()
		if err != nil {
			return err
		}
		mutex.Lock()
		s.data.Rank = ret
		mutex.Unlock()
		return nil
	})

	// 选股宝
	wg.Go(func() error {
		ret, err := s.xgb()
		if err != nil {
			return err
		}
		mutex.Lock()
		s.data.XGBMarketIndicator = ret
		mutex.Unlock()
		return nil
	})

	// xueqiu
	wg.Go(func() error {
		ret, err := s.xueqiu()
		if err != nil {
			return err
		}
		mutex.Lock()
		s.data.HotStock = ret
		mutex.Unlock()
		return nil
	})

	// 行业板块、概念版块数据
	wg.Go(func() error {
		industry, concept, err := s.sector()
		if err != nil {
			return err
		}
		mutex.Lock()
		s.data.IndustrySector = industry
		s.data.ConceptSector = concept
		mutex.Unlock()
		return nil
	})

	// 指数
	wg.Go(func() error {
		index, err := s.Index()
		if err != nil {
			return err
		}
		mutex.Lock()
		s.data.Index = index
		mutex.Unlock()
		return nil
	})
	if err := wg.Wait(); err != nil {
		return nil
	}
	// 没有问题则将s.data写入缓存
	s.data.UpdateTime = time.Now()
	buff, err := json.Marshal(s.data)
	if err != nil {
		return err
	}
	if _, err := db.Set(ctx, dataServiceCacheKey, buff, 7*24*time.Hour).Result(); err != nil {
		log.Errorf("RedisClient().Set(%+v) error: %+v", dataServiceCacheKey, err)
		return err
	}
	return nil
}

// Index 上证指数|深证指数|创业板指数
func (s *DataService) Index() ([]*model.Index, error) {
	codes := []string{"sh000001", "sz399001", "sz399006"}
	quoteURL := "http://qt.gtimg.cn/q="
	for _, it := range codes {
		quoteURL = fmt.Sprintf("%v,%v", quoteURL, it)
	}
	data, err := util.Http(quoteURL)
	if err != nil {
		return nil, serr.ErrBusiness("行情请求失败")
	}
	result := make([]*model.Index, 0)
	for -1 != strings.Index(data, "=") {
		item := &model.Index{}
		// 获取单条股票数据
		data = data[strings.Index(data, "\"")+1:]
		src := data[:strings.Index(data, "\"")]
		data = data[strings.Index(data, ";"):]

		// 分割字符串
		arr := strings.Split(src, "~")
		for index, value := range arr {
			switch index {
			case 1: // 指数名称
				item.Name = mahonia.NewDecoder("gbk").ConvertString(value)
			case 2: // 指数代码
				item.Code = value
			case 3: // 目前价格(指数)
				item.Price = util.String2Float64(value)
			case 31: // 涨跌额
				item.Chg = util.String2Float64(value)
			case 32: // 涨跌幅
				item.ChgPercent = util.String2Float64(value) / 100
			}
		}
		result = append(result, item)
	}
	return result, nil
}

// sector 获取行业版块、概念版块
func (s *DataService) sector() ([]*model.Sector, []*model.Sector, error) {
	type t struct {
		Data struct {
			Diff []struct {
				SectorName       string  `json:"f14"`  // 板块名称
				SectorCode       string  `json:"f12"`  // 板块代码
				SectorChgPercent float64 `json:"f3"`   // 版块涨跌幅度
				UpAmount         int64   `json:"f104"` // 上涨家数
				DownAmount       int64   `json:"f105"` // 下跌家数
				Turnover         float64 `json:"f8"`   // 版块换手率
				StockCode        string  `json:"f140"` // 股票代码
				StockName        string  `json:"f128"` // 股票名称
				StockChgPercent  float64 `json:"f136"` // 股票涨跌幅
			} `json:"diff"`
		} `json:"data"`
	}
	var industry, concept []*model.Sector
	for _, sectorUrl := range []string{industrySector, conceptSector} {
		url := sectorUrl
		response, err := util.Http(url)
		if err != nil {
			return nil, nil, err
		}
		sector := &t{}
		if err := json.Unmarshal([]byte(response), sector); err != nil {
			log.Errorf("Unmarshal err:%+v", err)
			return nil, nil, err
		}
		list := make([]*model.Sector, 0)
		for _, it := range sector.Data.Diff {
			list = append(list, &model.Sector{
				Code:            it.SectorCode,       // 板块代码
				Name:            it.SectorName,       // 板块名称
				ChgPercent:      it.SectorChgPercent, // 板块涨跌幅度
				UpAmount:        it.UpAmount,         // 上涨家数
				DownAmount:      it.DownAmount,       // 下跌家数
				Turnover:        it.Turnover,         // 版块换手率
				StockCode:       it.StockCode,        // 股票代码
				StockName:       it.StockName,        // 股票名称
				StockChgPercent: it.StockChgPercent,  // 股票涨跌幅度
				//StockPrice          				  // 股票价格-接口不返回股票价格
			})
		}
		switch url {
		case industrySector:
			industry = list
		case conceptSector:
			concept = list
		}
	}
	return industry, concept, nil
}

// xueqiu 雪球热门股票
func (s *DataService) xueqiu() ([]*model.HotStock, error) {
	// 创建fetcher
	newFetcher := func() *colly.Collector {
		fetcher := colly.NewCollector()
		fetcher.AllowURLRevisit = true
		extensions.Referer(fetcher)
		extensions.RandomUserAgent(fetcher)
		return fetcher
	}
	xueqiuCookie := func() ([]*http.Cookie, error) {
		fetcher := newFetcher()
		var cookie []*http.Cookie
		fetcher.OnResponse(func(response *colly.Response) {
			cookie = fetcher.Cookies(response.Request.URL.String())
		})
		if err := fetcher.Visit("https://xueqiu.com/"); err != nil {
			log.Errorf("访问雪球失败")
			return nil, err
		}
		return cookie, nil
	}
	cookie, err := xueqiuCookie()
	if err != nil {
		return nil, err
	}

	result := make([]*model.HotStock, 0)
	fetcher := newFetcher()
	fetcher.OnResponse(func(response *colly.Response) {
		type t struct {
			Data struct {
				Items []struct {
					Type      int64   `json:"type"`
					Code      string  `json:"code"`
					Name      string  `json:"name"`
					Value     float64 `json:"value"`
					Increment int64   `json:"increment"`
					Percent   float64 `json:"percent"`
					Current   float64 `json:"current"`
					Chg       float64 `json:"chg"`
				} `json:"items"`
			} `json:"data"`
		}
		xueqiu := &t{}
		if err := json.Unmarshal(response.Body, xueqiu); err != nil {
			log.Errorf("Unmarshal err:%+v", err)
		}
		for _, it := range xueqiu.Data.Items {
			result = append(result, &model.HotStock{
				Code:    it.Code[2:], // 股票代码,去掉SZ|SH
				Name:    it.Name,     // 股票名称
				Price:   it.Current,  // 股票价格
				Percent: it.Percent,  // 涨跌幅
				Chg:     it.Chg,      // 涨跌额
			})
		}
	})

	// 获取雪球热股
	hotStock := "https://stock.xueqiu.com/v5/stock/hot_stock/list.json?size=100&_type=12&type=12"
	if err := fetcher.SetCookies(hotStock, cookie); err != nil {
		log.Errorf("设置cookie失败")
		return nil, err
	}
	if err := fetcher.Visit(hotStock); err != nil {
		log.Errorf("Visit err:%+v", err)
		return nil, err
	}
	return result, nil
}

// xgb 选股宝市场指标
func (s *DataService) xgb() (*model.XGBMarketIndicator, error) {
	xgbMarketIndicatorUrl := "https://flash-api.xuangubao.cn/api/market_indicator/line?fields=rise_count,fall_count,stay_count,limit_up_count,limit_down_count,limit_up_broken_count,limit_up_broken_ratio,yesterday_limit_up_avg_pcp,market_temperature"
	resp, err := util.Http(xgbMarketIndicatorUrl)
	if err != nil {
		return nil, err
	}
	type t struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			FallCount              int64   `json:"fall_count"`
			LimitDownCount         int64   `json:"limit_down_count"`
			LimitUpBrokenCount     int64   `json:"limit_up_broken_count"`
			LimitUpBrokenRatio     float64 `json:"limit_up_broken_ratio"`
			LimitUpCount           int64   `json:"limit_up_count"`
			MarketTemperature      float64 `json:"market_temperature"`
			RiseCount              int64   `json:"rise_count"`
			StayCount              int64   `json:"stay_count"`
			Timestamp              int64   `json:"timestamp"`
			YesterdayLimitUpAvgPcp float64 `json:"yesterday_limit_up_avg_pcp"`
		} `json:"data"`
	}
	xgb := &t{}
	if err := json.Unmarshal([]byte(resp), xgb); err != nil {
		log.Errorf("Unmarshal err:%+v", err)
		return nil, err
	}
	if len(xgb.Data) == 0 {
		log.Errorf("结构体长度为0")
		return nil, serr.ErrBusiness("获取选股宝失败,xgb长度为0")
	}
	last := xgb.Data[len(xgb.Data)-1]
	return &model.XGBMarketIndicator{
		LimitDownCount:         last.LimitDownCount,
		LimitUpBrokenCount:     last.LimitUpBrokenCount,
		LimitUpBrokenRatio:     util.FloatRound(last.LimitUpBrokenRatio, 4),
		LimitUpCount:           last.LimitUpCount,
		YesterdayLimitUpAvgPcp: util.FloatRound(last.YesterdayLimitUpAvgPcp, 4),
	}, nil
}

// stockIndicatorRank 股票市场
func (s *DataService) stockIndicatorRank() (*model.Rank, error) {
	var mutex sync.Mutex
	rank := &model.Rank{}
	wg := errgroup.GroupWithCount(5)
	for _, rankUrl := range []string{upRankURL, downRankURL, turnoverRankURL, volumeRankURL, netInBalanceRankURL} {
		url := rankUrl
		wg.Go(func() error {
			ret, err := s.loadRank(url)
			if err != nil {
				return err
			}
			mutex.Lock()
			switch url {
			case upRankURL:
				rank.UpRank = ret
			case downRankURL:
				rank.DownRank = ret
			case turnoverRankURL:
				rank.TurnoverRank = ret
			case volumeRankURL:
				rank.VolumeRank = ret
			case netInBalanceRankURL:
				rank.NetInBalanceRank = ret
			}
			mutex.Unlock()
			return nil
		})
	}
	return rank, nil
}

// loadRank 获取股票排名
func (s *DataService) loadRank(url string) ([]*model.StockItem, error) {
	result := make([]*model.StockItem, 0)
	data, err := s.httpReq(url)
	if err != nil {
		return nil, err
	}
	// 解析
	for _, it := range data {
		item := &model.StockItem{}
		m := it.(map[string]interface{})
		for k, v := range m {
			switch k {
			case "f14": // 股票名称
				item.Name = v.(string)
			case "f12": // 股票代码
				item.Code = v.(string)
			case "f2": // 当前价格
				item.Price = v.(float64)
			case "f3": // 涨跌幅
				item.ChgPercent = v.(float64)
			}
		}
		// 过滤st|退市股票
		if strings.Contains(item.Name, "退") || strings.Contains(item.Name, "ST") || strings.Contains(item.Name, "N") || strings.Contains(item.Name, "C") {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

// getOverview 市场概况
func (s *DataService) getOverview() (*model.MarketOverView, error) {
	overviewMap, err := s.loadOverviewFromHttp()
	if err != nil {
		return nil, err
	}
	fn := func(overviewMap map[int64]int64, chg []int64) int64 {
		var sum int64
		for _, it := range chg {
			count, ok := overviewMap[it]
			if !ok {
				continue
			}
			sum += count
		}
		return sum
	}
	items := make([]*model.OverViewItem, 0) // 	// 涨停 >7 7-5 5-2 2-0 平 0-2 2-5 5-7 >7 跌停
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{11, 10}), Desc: "涨停"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{9, 8}), Desc: ">7"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{7, 6}), Desc: "7-5"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{5, 4, 3}), Desc: "5-2"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{2, 1}), Desc: "2-0"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{0}), Desc: "平"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{-2, -1}), Desc: "0-2"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{-5, -4, -3}), Desc: "2-5"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{-7, -6}), Desc: "5-7"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{-9, -8}), Desc: "<7"})
	items = append(items, &model.OverViewItem{Amount: fn(overviewMap, []int64{-11, -10}), Desc: "跌停"})
	var upAmount, downAmount, flatAmount int64 // 上涨，下跌，平 公司数
	for k, v := range overviewMap {
		if k > 0 {
			upAmount += v
		} else if k < 0 {
			downAmount += v
		} else {
			flatAmount += v
		}
	}
	return &model.MarketOverView{
		UpAmount:     upAmount,   // 上涨家数
		DownAmount:   downAmount, // 下跌家数
		FlatAmount:   flatAmount, // 平家数量
		OverViewItem: items,      // 市场概况分布
	}, nil
}

// loadOverviewFromHttp 根据http获取市场概况
func (s *DataService) loadOverviewFromHttp() (map[int64]int64, error) {
	result := make(map[int64]int64, 0)
	data, err := s.httpReq(overViewURL)
	if err != nil {
		return nil, err
	}
	for _, it := range data {
		m := it.(map[string]interface{})
		for k, v := range m {
			key, err := strconv.ParseInt(k, 10, 64)
			if err != nil {
				return nil, err
			}
			result[key] = int64(v.(float64))
		}
	}
	return result, nil
}

// httpReq http请求
func (s *DataService) httpReq(url string) ([]interface{}, error) {
	resp, err := util.Http(url)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(resp, "[") || !strings.Contains(resp, "]") {
		return nil, serr.ErrBusiness("代码不存在")
	}
	resp = resp[strings.Index(resp, "[") : strings.Index(resp, "]")+1]
	var data []interface{}
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
