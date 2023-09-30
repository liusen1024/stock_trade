package handler

import (
	"encoding/json"
	"fmt"
	"stock/api-gateway/quote"
	"stock/api-gateway/serr"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
	"stock/common/log"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"

	"github.com/gin-gonic/gin"
)

// HQHandler 行情
type HQHandler struct {
}

// NewHQHandler 单例
func NewHQHandler() *HQHandler {
	return &HQHandler{}
}

// Register 注册handler
func (h *HQHandler) Register(e *gin.Engine) {
	// 股票信息数据
	e.GET("/hq/stock_info", JSONWrapper(h.StockInfo)) // -->/h5_stockinfo
	// 分时
	e.GET("/hq/day", JSONWrapper(h.GetDay)) // --> /h5_fenshi
	// 5日k线
	e.GET("/hq/5day", PureJSONWrapper(h.Get5Day)) // -->/h5_fday
	// 日k线数据
	e.GET("/hq/kline/day", JSONWrapper(h.GetKlineDay)) // h5_dayline
	// 周k线数据
	e.GET("/hq/kline/week", JSONWrapper(h.GetKlineWeek)) // h5_weekline
	// 月k线数据
	e.GET("/hq/kline/month", JSONWrapper(h.GetKlineMonth)) // 月k线数据
	e.GET("/hq/data", PureJSONWrapper(h.GetData))
}

// GetData 获取数据
func (h *HQHandler) GetData(c *gin.Context) (interface{}, error) {

	type request struct {
		StockCode string `form:"stockno" json:"stockno"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.StockCode) == 0 {
		return nil, serr.ErrBusiness("股票不存在")
	}

	url := fmt.Sprintf("https://mnews.gw.com.cn/wap/data/ipad/stock/%s/%s/%s/f10/f10.html?qsThemeSign=1&themeStyleVs=1",
		util.GetStockMarketType(req.StockCode), req.StockCode[len(req.StockCode)-2:], req.StockCode)
	return map[string]interface{}{
		"code": "200",
		"url":  url,
	}, nil

}

// GetKlineMonth 获取月k
func (h *HQHandler) GetKlineMonth(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		StockCode string `form:"stockno" json:"stockno"`
		Date      string `form:"date" json:"date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.StockCode) == 0 {
		return nil, serr.ErrBusiness("股票不存在")
	}
	if len(req.Date) == 0 {
		req.Date = "1666878334939"
	}
	if _, err := service.StockDataServiceInstance().GetStockDataByCode(ctx, req.StockCode); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=%s%s&begin=%s&period=month&type=before&count=-285&indicator=kline,ma", util.GetStockMarketType(req.StockCode), req.StockCode, req.Date)
	return getKline(url)
}

// GetKlineWeek 获取周k
func (h *HQHandler) GetKlineWeek(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		StockCode string `form:"stockno" json:"stockno"`
		Date      string `form:"date" json:"date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.StockCode) == 0 {
		return nil, serr.ErrBusiness("股票不存在")
	}
	if len(req.Date) == 0 {
		req.Date = "1666878334939"
	}
	if _, err := service.StockDataServiceInstance().GetStockDataByCode(ctx, req.StockCode); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=%s%s&begin=%s&period=week&type=before&count=-285&indicator=kline,ma", util.GetStockMarketType(req.StockCode), req.StockCode, req.Date)
	return getKline(url)
}

// GetKlineDay 获取日K
func (h *HQHandler) GetKlineDay(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		StockCode string `form:"stockno" json:"stockno"`
		Date      string `form:"date" json:"date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.StockCode) == 0 {
		return nil, serr.ErrBusiness("股票不存在")
	}
	if len(req.Date) == 0 {
		req.Date = "1666878334939"
	}
	if _, err := service.StockDataServiceInstance().GetStockDataByCode(ctx, req.StockCode); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://stock.xueqiu.com/v5/stock/chart/kline.json?symbol=%s%s&begin=%s&period=day&type=before&count=-285&indicator=kline,ma", util.GetStockMarketType(req.StockCode), req.StockCode, req.Date)
	return getKline(url)
}

// Get5Day 获取5日分时图
func (h *HQHandler) Get5Day(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		StockCode string `form:"stockno" json:"stockno"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.StockCode) == 0 {
		return nil, serr.ErrBusiness("股票不存在")
	}
	// 查询数据库是否存在该股票
	if _, err := service.StockDataServiceInstance().GetStockDataByCode(ctx, req.StockCode); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://stock.xueqiu.com/v5/stock/chart/minute.json?symbol=%s%s&period=5d", util.GetStockMarketType(req.StockCode), req.StockCode)
	return getMinuteKLine(url)
}

type T2 struct {
	Data struct {
		LastClose float64 `json:"last_close"`

		Items []struct {
			Current   float64     `json:"current"`
			Volume    int         `json:"volume"`
			AvgPrice  float64     `json:"avg_price"`
			Chg       float64     `json:"chg"`
			Percent   float64     `json:"percent"`
			Timestamp int64       `json:"timestamp"`
			Amount    float64     `json:"amount"`
			High      float64     `json:"high"`
			Low       float64     `json:"low"`
			Macd      interface{} `json:"macd"`
			Kdj       interface{} `json:"kdj"`
			Ratio     interface{} `json:"ratio"`
			Capital   *struct {
				Small  float64 `json:"small"`
				Medium float64 `json:"medium"`
				Large  float64 `json:"large"`
				Xlarge float64 `json:"xlarge"`
			} `json:"capital"`
			VolumeCompare struct {
				VolumeSum     int `json:"volume_sum"`
				VolumeSumLast int `json:"volume_sum_last"`
			} `json:"volume_compare"`
		} `json:"items"`
	} `json:"data"`
}

func getMinuteKLine(url string) (*MinuteKlineResp, error) {
	cookies, err := util.XueQiuCookie()
	if err != nil {
		return nil, err
	}
	fetcher := newFetcher()
	if err := fetcher.SetCookies(url, cookies); err != nil {
		log.Errorf("设置cookie失败,err:%+v", err)
		return nil, err
	}
	minuteKlines := make([]*MinuteKlineItem, 0)
	var lastClose float64
	fetcher.OnResponse(func(response *colly.Response) {
		type xueQiu struct {
			Data struct {
				LastClose float64            `json:"last_close"`
				Items     []*MinuteKlineItem `json:"items"`
			} `json:"data"`
		}
		data := &xueQiu{}
		if err := json.Unmarshal(response.Body, data); err != nil {
			log.Errorf("Unmarshal err:%+v", err)
		}
		for _, it := range data.Data.Items {
			minuteKlines = append(minuteKlines, it)
		}
		lastClose = data.Data.LastClose
	})
	if err := fetcher.Visit(url); err != nil {
		log.Errorf("Visit err:%+v", err)
		return nil, err
	}

	return &MinuteKlineResp{
		Code:      "100",
		Data:      minuteKlines,
		LastClose: lastClose,
	}, nil
}

// getKline 获取日K,周k，月k
func getKline(url string) (interface{}, error) {
	cookies, err := util.XueQiuCookie()
	if err != nil {
		return nil, err
	}
	fetcher := newFetcher()
	if err := fetcher.SetCookies(url, cookies); err != nil {
		log.Errorf("设置cookie失败,err:%+v", err)
		return nil, err
	}
	kline := make([][]float64, 0)
	fetcher.OnResponse(func(response *colly.Response) {
		type xueQiu struct {
			Data struct {
				Item [][]float64 `json:"item"`
			} `json:"data"`
		}
		data := &xueQiu{}
		if err := json.Unmarshal(response.Body, data); err != nil {
			log.Errorf("Unmarshal err:%+v", err)
		}
		for _, list := range data.Data.Item {
			a := list[0:10]
			b := list[12:]
			list = list[0:0]
			list = append(list, a...)
			list = append(list, b...)
			kline = append(kline, list)
		}
	})
	if err := fetcher.Visit(url); err != nil {
		log.Errorf("Visit err:%+v", err)
		return nil, err
	}
	return map[string]interface{}{
		"code": "100",
		"data": kline,
	}, nil
}

// GetDay 获取分时数据
func (h *HQHandler) GetDay(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		StockCode string `form:"stockno" json:"stockno"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.StockCode) == 0 {
		return nil, serr.ErrBusiness("股票不存在")
	}
	// 查询数据库是否存在该股票
	if _, err := service.StockDataServiceInstance().GetStockDataByCode(ctx, req.StockCode); err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://stock.xueqiu.com/v5/stock/chart/minute.json?symbol=%s%s&period=1d", util.GetStockMarketType(req.StockCode), req.StockCode)
	klines, err := getMinuteKLine(url)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"code": "100",
		"data": klines.Data,
	}, nil
}

// StockInfo 股票信息
func (h *HQHandler) StockInfo(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		StockCode string `form:"stockno" json:"stockno"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.StockCode) == 0 {
		return nil, serr.ErrBusiness("股票不存在")
	}

	type resp struct {
		Buyfivenum    string `json:"buyfivenum"`
		Buyfivepri    string `json:"buyfivepri"`
		Buyfournum    string `json:"buyfournum"`
		Buyfourpri    string `json:"buyfourpri"`
		Buyonenum     string `json:"buyonenum"`
		Buyonepri     string `json:"buyonepri"`
		Buythreenum   string `json:"buythreenum"`
		Buythreepri   string `json:"buythreepri"`
		Buytwonum     string `json:"buytwonum"`
		Buytwopri     string `json:"buytwopri"`
		Chg           string `json:"chg"`
		Code          string `json:"code"`
		IndexColor    string `json:"index_color"`
		Message       string `json:"message"`
		Pb            string `json:"pb"`
		Pe            string `json:"pe"`
		Percent       string `json:"percent"`
		Price         string `json:"price"`
		Sellfivenum   string `json:"sellfivenum"`
		Sellfivepri   string `json:"sellfivepri"`
		Sellfournum   string `json:"sellfournum"`
		Sellfourpri   string `json:"sellfourpri"`
		Sellonenum    string `json:"sellonenum"`
		Sellonepri    string `json:"sellonepri"`
		Sellthreenum  string `json:"sellthreenum"`
		Sellthreepri  string `json:"sellthreepri"`
		Selltwonum    string `json:"selltwonum"`
		Selltwopri    string `json:"selltwopri"`
		Stockname     string `json:"stockname"`
		Stockno       string `json:"stockno"`
		Todaymax      string `json:"todaymax"`
		Todaymin      string `json:"todaymin"`
		Todaystartpri string `json:"todaystartpri"`
		Traamount     string `json:"traamount"`
		Tranumber     string `json:"tranumber"`
		Yestodendpri  string `json:"yestodendpri"`
	}
	// 查询数据库是否存在该股票
	if _, err := service.StockDataServiceInstance().GetStockDataByCode(ctx, req.StockCode); err != nil {
		return nil, err
	}

	qt, err := quote.QtServiceInstance().GetQuoteByTencent([]string{req.StockCode})
	if err != nil {
		return nil, err
	}
	stock, ok := qt[req.StockCode]
	if !ok {
		return nil, serr.ErrBusiness("证券代码不存在")
	}
	indexColor := "0"
	if stock.Chg > 0 {
		indexColor = "1"
	}
	return resp{
		Buyfivenum:    fmt.Sprintf("%v", stock.BuyVol5),
		Buyfivepri:    fmt.Sprintf("%v", stock.BuyPrice5),
		Buyfournum:    fmt.Sprintf("%v", stock.BuyVol4),
		Buyfourpri:    fmt.Sprintf("%v", stock.BuyPrice4),
		Buyonenum:     fmt.Sprintf("%v", stock.BuyVol1),
		Buyonepri:     fmt.Sprintf("%v", stock.BuyPrice1),
		Buythreenum:   fmt.Sprintf("%v", stock.BuyVol3),
		Buythreepri:   fmt.Sprintf("%v", stock.BuyPrice3),
		Buytwonum:     fmt.Sprintf("%v", stock.BuyVol2),
		Buytwopri:     fmt.Sprintf("%v", stock.BuyPrice2),
		Chg:           util.WithSigned(stock.Chg),
		Code:          "100",      // 状态:100 表示成功
		IndexColor:    indexColor, // 颜色:0绿色 1红色
		Message:       "",
		Pb:            fmt.Sprintf("%v", stock.Pb),
		Pe:            fmt.Sprintf("%v", stock.Pe),
		Percent:       util.WithPercent(stock.ChgPercent),
		Price:         fmt.Sprintf("%v", stock.CurrentPrice),
		Sellfivenum:   fmt.Sprintf("%v", stock.SellVol5),
		Sellfivepri:   fmt.Sprintf("%v", stock.SellPrice5),
		Sellfournum:   fmt.Sprintf("%v", stock.SellVol4),
		Sellfourpri:   fmt.Sprintf("%v", stock.SellPrice4),
		Sellonenum:    fmt.Sprintf("%v", stock.SellVol1),
		Sellonepri:    fmt.Sprintf("%v", stock.SellPrice1),
		Sellthreenum:  fmt.Sprintf("%v", stock.SellVol3),
		Sellthreepri:  fmt.Sprintf("%v", stock.SellPrice3),
		Selltwonum:    fmt.Sprintf("%v", stock.SellVol2),
		Selltwopri:    fmt.Sprintf("%v", stock.SellPrice2),
		Stockname:     stock.Name,
		Stockno:       stock.Code,
		Todaymax:      fmt.Sprintf("%v", stock.HighPx),
		Todaymin:      fmt.Sprintf("%v", stock.LowPx),
		Todaystartpri: fmt.Sprintf("%v", stock.OpenPrice),
		Traamount:     fmt.Sprintf("%v", stock.TotalAmount),
		Tranumber:     fmt.Sprintf("%v", stock.TotalVol),
		Yestodendpri:  fmt.Sprintf("%v", stock.ClosePrice),
	}, nil
}

func newFetcher() *colly.Collector {
	fetcher := colly.NewCollector()
	fetcher.AllowURLRevisit = true
	extensions.Referer(fetcher)
	extensions.RandomUserAgent(fetcher)
	return fetcher
}
