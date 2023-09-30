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
	"stock/common/timeconv"
	"sync"
	"time"
)

// PositionService 持仓服务
type PositionService struct {
}

var (
	positionService *PositionService
	positionOnce    sync.Once
)

// PositionServiceInstance PositionServiceInstance实例
func PositionServiceInstance() *PositionService {
	positionOnce.Do(func() {
		positionService = &PositionService{}

		ctx := context.Background()

		// 解冻所有的股票
		go func() {
			for range time.Tick(5 * time.Second) {
				if !CalendarServiceInstance().IsTradeDate(ctx) || time.Now().Hour() < 16 || time.Now().Minute() < 30 {
					continue
				}
				// 检查redis是否已经执行过
				key := fmt.Sprintf("position_unfreeze_amount_task_%v", timeconv.TimeToInt32(time.Now()))
				if db.RedisClient().Get(ctx, key).Val() == "1" {
					continue
				}
				if err := positionService.unFreezeStockTask(ctx); err != nil {
					log.Errorf("unFreezeStock err:%+v", err)
					continue
				}
				if err := db.RedisClient().Set(ctx, key, "1", -1).Err(); err != nil {
					log.Errorf("设置自动撤单失败:%+v", err)
				}
			}
		}()

	})
	return positionService
}

// unFreezeStockTask 交易日16:30解冻股票
func (s *PositionService) unFreezeStockTask(ctx context.Context) error {
	list, err := dao.PositionDaoInstance().GetPositions(ctx)
	if err != nil {
		log.Errorf("查询持仓失败:%+v", err)
		return nil
	}
	if len(list) == 0 {
		return nil
	}

	for _, item := range list {
		position := item
		position.FreezeAmount = 0
		if err := dao.PositionDaoInstance().Update(ctx, position); err != nil {
			log.Errorf("解冻股票失败:%+v", err)
			return nil
		}
	}
	log.Infof("解冻股数任务完成!")
	return nil
}

// getPositionQuote 查询持仓数据刷新行情缓存
func (s *PositionService) getPositionQuote(ctx context.Context) error {
	position, err := dao.PositionDaoInstance().GetPositions(ctx)
	if err != nil {
		return err
	}
	codes := make([]string, 0)
	for _, it := range position {
		codes = append(codes, it.StockCode)
	}
	if _, err := quote.QtServiceInstance().GetQuoteByTencent(codes); err != nil {
		return err
	}
	return nil
}

// GetPositionByContractID 根据合约ID查询持仓
func (s *PositionService) GetPositionByContractID(ctx context.Context, contractID int64) ([]*model.Position, error) {
	positions, err := dao.PositionDaoInstance().GetPositionByContractID(ctx, contractID)
	if err != nil {
		log.Errorf("GetPositionByContractID err:%+v", err)
		return nil, err
	}
	if len(positions) == 0 {
		return positions, nil
	}
	return s.setCurPrice(positions)
}

// GetPositionByEntrustID 通过持仓编号查询持仓记录
func (s *PositionService) GetPositionByEntrustID(ctx context.Context, entrustID int64) (*model.Position, error) {
	position, err := dao.PositionDaoInstance().GetPositionByEntrustID(ctx, entrustID)
	if err != nil {
		log.Errorf("GetPositionByEntrustID err:%+v", err)
		return nil, err
	}
	result, err := s.setCurPrice([]*model.Position{position})
	if err != nil {
		log.Errorf("setCurPrice err:%+v", err)
		return nil, err
	}
	if len(result) > 0 {
		return result[0], nil
	}
	return nil, serr.ErrBusiness("持仓记录不存在")
}

// setCurPrice 设置当前价格
func (s *PositionService) setCurPrice(positions []*model.Position) ([]*model.Position, error) {
	codes := make([]string, 0)
	for _, it := range positions {
		codes = append(codes, it.StockCode)
	}
	qts, err := quote.QtServiceInstance().GetQuoteByTencent(codes)
	if err != nil {
		log.Errorf("GetQuoteByTencent err:%+v", err)
		return nil, err
	}
	for _, it := range positions {
		qt, ok := qts[it.StockCode]
		if !ok {
			log.Errorf("获取持仓行情失败:code[%+v]", it.StockCode)
			continue
		}
		it.CurPrice = qt.CurrentPrice
	}
	return positions, nil
}
