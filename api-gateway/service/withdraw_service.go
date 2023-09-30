package service

import (
	"context"
	"fmt"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/common/log"
	"stock/common/timeconv"
	"sync"
	"time"
)

// WithdrawService 服务
type WithdrawService struct {
}

var (
	withdrawService *WithdrawService
	withdrawOnce    sync.Once
)

// WithdrawServiceInstance 实例
func WithdrawServiceInstance() *WithdrawService {
	withdrawOnce.Do(func() {
		withdrawService = &WithdrawService{}
		ctx := context.Background()

		// 未成交则自动撤单
		go func() {
			for range time.Tick(5 * time.Second) {
				if !CalendarServiceInstance().IsTradeDate(ctx) || time.Now().Hour() <= 15 {
					continue
				}

				// 检查redis是否已经执行过
				key := fmt.Sprintf("auto_withdraw_key_%v", timeconv.TimeToInt32(time.Now()))
				if db.RedisClient().Get(ctx, key).Val() == "1" {
					continue
				}
				if err := withdrawService.withdrawTask(ctx); err != nil {
					log.Errorf("withdrawTask err:%+v", err)
				}
				if err := db.RedisClient().Set(ctx, key, "1", 2*24*time.Hour).Err(); err != nil {
					log.Errorf("设置自动撤单失败:%+v", err)
				}
				log.Infof("系统任务,自动撤单成功!")
			}
		}()
	})
	return withdrawService
}

// withdrawTask 未成交自动撤单
func (s *WithdrawService) withdrawTask(ctx context.Context) error {
	list, err := dao.EntrustDaoInstance().GetTodayEntrusts(ctx)
	if err != nil {
		log.Errorf("GetTodayEntrusts err:%+v", err)
		return err
	}
	if len(list) == 0 {
		return nil
	}
	for _, entrust := range list {
		if entrust.IsFinallyState() {
			continue
		}
		entrust.Status = model.EntrustStatusTypeWithdraw
		if err := dao.EntrustDaoInstance().Update(ctx, entrust); err != nil {
			return err
		}

		// 券商委托表填写撤单
		if entrust.IsBrokerEntrust {
			if err := s.brokerEntrustWithdraw(ctx, entrust); err != nil {
				log.Errorf("brokerEntrustWithdraw err:%+v", err)
				return err
			}
		}

		// 更新可用资金
		if err := ContractServiceInstance().UpdateValMoneyByID(ctx, entrust.ContractID); err != nil {
			log.Errorf("UpdateValMoneyByID err:%+v", err)
		}
	}
	return nil
}

// brokerEntrustWithdraw 填写券商委托表撤单
func (s *WithdrawService) brokerEntrustWithdraw(ctx context.Context, entrust *model.Entrust) error {
	brokerEntrusts, err := dao.BrokerEntrustDaoInstance().GetByEntrustID(ctx, entrust.ID)
	if err != nil {
		return err
	}
	if len(brokerEntrusts) == 0 {
		return nil
	}
	for _, it := range brokerEntrusts {
		if !it.IsFinallyState() {
			it.Status = model.EntrustStatusTypeWithdraw
		}
	}
	if err := dao.BrokerEntrustDaoInstance().MCreate(ctx, brokerEntrusts); err != nil {
		log.Errorf("券商委托表填写撤单失败:%+v", err)
	}
	return nil
}
