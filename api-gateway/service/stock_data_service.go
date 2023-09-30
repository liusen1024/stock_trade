package service

import (
	"context"
	"encoding/json"
	"fmt"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/util"
	"stock/common/log"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

// StockDataService 服务
type StockDataService struct {
}

var (
	stockDataService *StockDataService
	stockDataOnce    sync.Once
)

// StockDataServiceInstance StockDataServiceInstance实例
func StockDataServiceInstance() *StockDataService {
	stockDataOnce.Do(func() {
		stockDataService = &StockDataService{}
		ctx := context.Background()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("stockData service get panic:%+v", r)
				}
			}()
			for range time.Tick(12 * time.Hour) {
				if err := stockDataService.get(ctx); err != nil {
					log.Errorf("刷新股票池失败:%+v", err)
					continue
				}
			}
		}()
	})
	return stockDataService
}

// stockCacheKey 股票缓存key
func (s *StockDataService) stockCacheKey() string {
	return "load_stock_data_list"
}

// get 从网络获取股票信息
func (s *StockDataService) get(ctx context.Context) error {
	list := make([]*model.StockData, 0)
	getStockData := func(url string) error {
		fetcher := colly.NewCollector()
		fetcher.AllowURLRevisit = true
		extensions.Referer(fetcher)
		extensions.RandomUserAgent(fetcher)
		fetcher.OnResponse(func(response *colly.Response) {
			type t struct {
				Result struct {
					Data []struct {
						SECURITYCODE    string  `json:"SECURITY_CODE"`
						SECURITYNAME    string  `json:"SECURITY_NAME"`
						TRADEMARKETCODE string  `json:"TRADE_MARKET_CODE"`
						APPLYCODE       *string `json:"APPLY_CODE"`
						TRADEMARKET     string  `json:"TRADE_MARKET"`
						MARKETTYPE      string  `json:"MARKET_TYPE"`
						ORGTYPE         string  `json:"ORG_TYPE"`
						LISTINGDATE     *string `json:"LISTING_DATE"`
						ISBEIJING       int     `json:"IS_BEIJING"`
					} `json:"data"`
				} `json:"result"`
				Code int `json:"code"`
			}
			d := &t{}
			if err := json.Unmarshal(response.Body, d); err != nil {
				log.Errorf("Unmarshal err:%+v", err)
				return
			}
			if d.Code != 0 {
				log.Errorf("返回码错误:%+v", d.Code)
				return
			}
			for _, it := range d.Result.Data {
				// 过滤B股
				if strings.Contains(it.SECURITYNAME, "B") {
					continue
				}
				ipoTime := time.Now()
				if it.LISTINGDATE != nil {
					date, err := time.Parse("2006-01-02 15:04:05", *it.LISTINGDATE)
					if err == nil {
						ipoTime = date
					}
				}
				list = append(list, &model.StockData{
					Code:   it.SECURITYCODE, // 股票代码
					Name:   it.SECURITYNAME, // 股票名称
					IPODay: ipoTime,         // IPO日期
					Status: model.StockDataStatusEnable,
				})
			}
		})
		if err := fetcher.Visit(url); err != nil {
			return err
		}
		return nil
	}

	for page := 1; page < 20; page++ {
		p := page
		url := fmt.Sprintf("https://datacenter-web.eastmoney.com/api/data/v1/get?sortColumns=APPLY_DATE,SECURITY_CODE&pageSize=500&pageNumber=%+v&reportName=RPTA_APP_IPOAPPLY&columns=SECURITY_CODE,SECURITY_NAME,TRADE_MARKET_CODE,APPLY_CODE,TRADE_MARKET,MARKET_TYPE,ORG_TYPE,LISTING_DATE,IS_BEIJING,INDUSTRY_PE_RATIO&quoteColumns=f2~01~SECURITY_CODE~NEWEST_PRICE&quoteType=0&source=WEB&client=WEB", p)
		if err := getStockData(url); err != nil {
			log.Errorf("getStockData err:%+v", err)
			return err
		}
	}
	// 更新
	if err := dao.StockDataDaoInstance().Update(ctx, list); err != nil {
		log.Errorf("更新股票列表失败:%+v", err)
		return err
	}

	if err := db.RedisClient().Del(ctx, s.stockCacheKey()).Err(); err != nil {
		log.Errorf("删除redis key:%+v 失败,err:%+v", s.stockCacheKey(), err)
	}

	// 重新载入缓存
	if _, err := s.GetStocks(ctx); err != nil {
		log.Errorf("载入缓存失败:%+v", err)
	}
	return nil
}

// GetStockDataByCode 查询股票
func (s *StockDataService) GetStockDataByCode(ctx context.Context, code string) (*model.StockData, error) {
	stock, err := dao.StockDataDaoInstance().GetStockDataByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return stock, err
}

// IsTrade 是否可以买入
func (s *StockDataService) IsTrade(ctx context.Context, code string) error {
	stock, err := s.GetStockDataByCode(ctx, code)
	if err != nil {
		return err
	}
	if stock.Status == model.StockDataStatusDisable {
		return serr.New(serr.ErrCodeBusinessFail, "委托失败,[风控提示]该股票不可交易")
	}
	return nil
}

// GetStocks 获取可交易的股票列表
func (s *StockDataService) GetStocks(ctx context.Context) ([]*model.StockData, error) {
	var list []*model.StockData
	if err := db.GetOrLoad(ctx, s.stockCacheKey(), 6*time.Hour, &list, func() error {
		stockList, err := dao.StockDataDaoInstance().Get(ctx)
		if err != nil {
			return err
		}
		sys, err := dao.SysDaoInstance().GetSysParam(ctx)
		if err != nil {
			return err
		}
		// 是否允许交易
		for _, it := range stockList {
			switch util.StockBord(it.Code) {
			case util.StockTypeKCBBORD:
				if !sys.IsSupportKCBBoard {
					continue
				}
			case util.StockTypeBJ:
				if !sys.IsSupportBJBoard {
					continue
				}
			case util.StockTypeCYBBORD:
				if !sys.IsSupportCYBBoard {
					continue
				}
			}
			if it.Status == model.StockDataStatusEnable {
				list = append(list, it)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return list, nil
}

// Update 更新股票列表
func (s *StockDataService) Update(ctx context.Context, list []*model.StockData) error {
	return dao.StockDataDaoInstance().Update(ctx, list)
}

// UpdateStatusByID 根据ID更新股票交易状态
func (s *StockDataService) UpdateStatusByID(ctx context.Context, id int64, status int64) error {
	if err := dao.StockDataDaoInstance().UpdateStatusByID(ctx, id, status); err != nil {
		log.Errorf("UpdateStatusByID err:%+v", err)
		return err
	}
	if err := db.RedisClient().Del(ctx, s.stockCacheKey()).Err(); err != nil {
		log.Errorf("删除redis缓存失败,key:%+v err:%+v", s.stockCacheKey(), err)
		return err
	}
	return nil
}
