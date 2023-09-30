package service

import (
	"context"
	"fmt"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/quote"
	"stock/api-gateway/serr"
	"stock/common/log"
	"sync"
	"time"
)

// PortfolioService 服务
type PortfolioService struct {
}

var (
	portfolioService *PortfolioService
	portfolioOnce    sync.Once
)

// PortfolioServiceInstance 实例
func PortfolioServiceInstance() *PortfolioService {
	portfolioOnce.Do(func() {
		portfolioService = &PortfolioService{}
	})
	return portfolioService
}

// GetPortfolioList 查询自选股列表
func (s *PortfolioService) GetPortfolioList(ctx context.Context, uid int64) ([]*model.StockItem, error) {
	return s.getPortfolioList(ctx, uid)
}

// DeletePortfolio 删除自选股
func (s *PortfolioService) DeletePortfolio(ctx context.Context, uid int64, code string) ([]*model.StockItem, error) {
	if err := dao.PortfolioDaoInstance().DeletePortfolio(ctx, uid, code); err != nil {
		return nil, serr.ErrBusiness("删除自选股失败")
	}
	return s.getPortfolioList(ctx, uid)
}

// CreatePortfolio 新增自选股
func (s *PortfolioService) CreatePortfolio(ctx context.Context, uid int64, code string) ([]*model.StockItem, error) {
	qtMap, err := quote.QtServiceInstance().GetQuoteByTencent([]string{code})
	if err != nil {
		return nil, serr.ErrBusiness("获取行情失败")
	}
	qt, ok := qtMap[code]
	if !ok {
		return nil, serr.ErrBusiness("获取行情失败")
	}
	if err := dao.PortfolioDaoInstance().CreatePortfolio(ctx, &model.Portfolio{
		UID:        uid,
		StockCode:  code,
		StockName:  qt.Name,
		CreateTime: time.Now(),
	}); err != nil {
		return nil, serr.ErrBusiness("添加自选股失败")
	}
	return s.getPortfolioList(ctx, uid)
}

// getPortfolioList 查询自选股
func (s *PortfolioService) getPortfolioList(ctx context.Context, uid int64) ([]*model.StockItem, error) {
	loadFn := func(ctx context.Context, uid int64) ([]*model.StockItem, error) {
		result := make([]*model.StockItem, 0)
		list, err := dao.PortfolioDaoInstance().GetPortfolioList(ctx, uid)
		if err != nil {
			return nil, serr.ErrBusiness("查询自选股失败")
		}
		if len(list) == 0 {
			return result, nil
		}
		// 根据list股票查询股票行情
		codes := make([]string, 0, len(list))
		for _, it := range list {
			codes = append(codes, it.StockCode)
		}
		qtMap, err := quote.QtServiceInstance().GetQuoteByTencent(codes)
		if err != nil {
			return nil, serr.ErrBusiness("获取行情失败")
		}
		for _, it := range list {
			qt, ok := qtMap[it.StockCode]
			if !ok {
				continue
			}
			result = append(result, &model.StockItem{
				Code:       it.StockCode,
				Name:       it.StockName,
				Price:      qt.CurrentPrice,
				ChgPercent: qt.ChgPercent,
			})
		}
		return result, nil
	}

	// 缓存
	var list []*model.StockItem
	if err := db.GetOrLoad(ctx, fmt.Sprintf("portfolio_uid_%d_list", uid), 3*time.Second, &list, func() error {
		ret, err := loadFn(ctx, uid)
		if err != nil {
			log.Errorf("获取自选股失败:%+v", err)
			return err
		}
		list = ret
		return nil
	}); err != nil {
		return nil, serr.ErrBusiness("获取自选股列表失败")
	}
	return list, nil
}
