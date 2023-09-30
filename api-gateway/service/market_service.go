package service

import (
	"context"
	"encoding/json"
	"fmt"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/quote"
	"stock/api-gateway/serr"
	"stock/api-gateway/util"
	"stock/common/log"
	"strings"
	"sync"
	"time"
)

// MarketService 服务
type MarketService struct {
}

var (
	marketService *MarketService
	marketOnce    sync.Once
)

// MarketServiceInstance 实例
func MarketServiceInstance() *MarketService {
	marketOnce.Do(func() {
		marketService = &MarketService{}
	})
	return marketService
}

// GetMarket 市场界面
func (s *MarketService) GetMarket(ctx context.Context) (*model.Market, error) {
	var market *model.Market
	if err := db.GetOrLoad(ctx, fmt.Sprintf("market"), time.Second*3, &market, func() error {
		ret, err := s.getMarket(ctx)
		if err != nil {
			return err
		}
		market = ret
		return nil
	}); err != nil {
		return nil, err
	}
	return market, nil
}

// GetBKDetail 获取某个行业的全部股票信息
func (s *MarketService) GetBKDetail(ctx context.Context, code string) ([]*model.StockItem, error) {
	list, err := s.getSectorByBKNo(code)
	if err != nil {
		log.Errorf("getSectorByBKNo err :%+v 板块代码:%+v", err, code)
		return nil, serr.ErrBusiness("获取板块信息失败")
	}
	return list, nil
}

func (s *MarketService) getSectorByBKNo(code string) ([]*model.StockItem, error) {
	url := fmt.Sprintf("https://push2.eastmoney.com/api/qt/clist/get?fid=f62&po=1&pz=200&pn=1&np=1&fltt=2&invt=2&fs=b:%s&fields=f12,f14,f2,f3", code)
	resp, err := util.Http(url)
	if err != nil {
		return nil, err
	}
	type t struct {
		Data struct {
			List []struct {
				Price      float64 `json:"f2"`  // 当前价格
				ChgPercent float64 `json:"f3"`  // 涨跌幅
				Code       string  `json:"f12"` // 股票代码
				Name       string  `json:"f14"` // 股票名称
			} `json:"diff"`
		} `json:"data"`
	}
	items := &t{}
	if err := json.Unmarshal([]byte(resp), items); err != nil {
		return nil, err
	}
	result := make([]*model.StockItem, 0)
	for _, it := range items.Data.List {
		// 过滤st|退市股票
		if filter(it.Name) || it.Price == 0 {
			continue
		}
		result = append(result, &model.StockItem{
			Price:      it.Price,
			ChgPercent: it.ChgPercent,
			Code:       it.Code,
			Name:       it.Name,
		})
	}
	return result, nil
}

// filter 过滤股票规则:当返回true 表示需要过滤
func filter(stockName string) bool {
	if strings.Contains(stockName, "退") || strings.Contains(stockName, "ST") ||
		strings.Contains(stockName, "B") || strings.Contains(stockName, "N") {
		return true
	}
	return false
}

// GetAllIndustry 获取所有的行业
func (s *MarketService) GetAllIndustry(ctx context.Context) ([]*model.Sector, error) {
	data := DataServiceInstance().GetData()
	return s.getAllSector(data.IndustrySector)
}

func (s *MarketService) getAllSector(sector []*model.Sector) ([]*model.Sector, error) {
	codes := make([]string, 0)
	for _, industry := range sector {
		codes = append(codes, industry.StockCode)
	}
	// GetData()数据不包含股票价格,因而需要去重新获取
	qtMap, err := quote.QtServiceInstance().GetQuoteByTencent(codes)
	if err != nil {
		return nil, err
	}
	for _, it := range sector {
		qt, ok := qtMap[it.StockCode]
		if !ok {
			continue
		}
		it.StockPrice = qt.CurrentPrice
	}
	return sector, nil
}

// GetAllConcept 获取所有的概念板块
func (s *MarketService) GetAllConcept(ctx context.Context) ([]*model.Sector, error) {
	data := DataServiceInstance().GetData()
	return s.getAllSector(data.ConceptSector)
}

// getMarket 获取市场
func (s *MarketService) getMarket(ctx context.Context) (*model.Market, error) {
	data := DataServiceInstance().GetData()
	temperature := 0.5
	if data.MarketOverView.UpAmount != 0 && data.MarketOverView.DownAmount != 0 {
		temperature = float64(data.MarketOverView.UpAmount) / float64(data.MarketOverView.UpAmount+data.MarketOverView.DownAmount)
	}
	codes := make([]string, 0)
	for _, it := range data.IndustrySector {
		codes = append(codes, it.StockCode)
	}
	for _, it := range data.ConceptSector {
		codes = append(codes, it.StockCode)
	}
	qtMap, err := quote.QtServiceInstance().GetQuoteByTencent(codes)
	if err != nil {
		log.Errorf("GetQuoteByTencent err:%+v", err)
		return nil, serr.ErrBusiness("获取市场失败")
	}
	result := &model.Market{
		Index:             data.Index,
		OverViewItem:      data.MarketOverView.OverViewItem,
		UpAmount:          data.MarketOverView.UpAmount,
		DownAmount:        data.MarketOverView.DownAmount,
		FlatAmount:        data.MarketOverView.FlatAmount,
		MarketTemperature: temperature,
		Industry:          s.setPrice(data.IndustrySector, qtMap),
		Concept:           s.setPrice(data.ConceptSector, qtMap),
		UpRank:            data.Rank.UpRank,
		DownRank:          data.Rank.DownRank,
		TurnoverRank:      data.Rank.TurnoverRank,
		VolumeRank:        data.Rank.VolumeRank,
		NetInBalanceRank:  data.Rank.NetInBalanceRank,
	}
	return result, nil
}

// setPrice 设置板块股票的价格
func (s *MarketService) setPrice(list []*model.Sector, qtMap map[string]*model.TencentQuote) []*model.Sector {
	result := make([]*model.Sector, 0)
	for _, sector := range list {
		qt, ok := qtMap[sector.StockCode]
		if !ok {
			continue
		}
		result = append(result, &model.Sector{
			Code:            sector.Code,            // 板块代码
			Name:            sector.Name,            // 板块名称
			ChgPercent:      sector.ChgPercent,      // 板块涨跌幅度
			UpAmount:        sector.UpAmount,        // 上涨家数
			DownAmount:      sector.DownAmount,      // 下跌家数
			Turnover:        sector.Turnover,        // 版块换手率
			StockCode:       sector.StockCode,       // 股票代码
			StockName:       sector.StockName,       // 股票名称
			StockPrice:      qt.CurrentPrice,        // 股票价格
			StockChgPercent: sector.StockChgPercent, // 股票涨跌幅度
		})
	}
	return result
}
