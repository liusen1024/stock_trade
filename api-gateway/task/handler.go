package service

import (
	"context"
	"stock/api-gateway/dao"
	"stock/api-gateway/model"
	"stock/common/log"
	"time"
)

// HisTrade 历史交易数据归档
func HisTrade() {
	ctx := context.Background()
	positions, err := dao.PositionDaoInstance().GetPositions(ctx)
	if err != nil {
		log.Errorf("历史数据归档失败:查询持仓表失败")
	}
	if len(positions) == 0 {
		log.Info("无持仓记录归档")
		return
	}
	list := make([]*model.Position, 0)
	for _, it := range positions {
		list = append(list, &model.Position{
			UID:          it.UID,
			ContractID:   it.ContractID,
			OrderTime:    time.Now(),
			StockCode:    it.StockCode,
			StockName:    it.StockName,
			Price:        it.Price,
			Amount:       it.Amount,
			Balance:      it.Balance,
			FreezeAmount: it.FreezeAmount,
		})
	}
	if err := dao.HisPositionDaoInstance().Create(ctx, list); err != nil {
		log.Error("历史数据归档失败")
	}
	log.Infof("历史数据归档完毕!")
}

// ContractRecord 每天23:00分执行一次 归档本日合约
func ContractRecord() {
	ctx := context.Background()
	if err := dao.ContractRecordDaoInstance().Copy(ctx); err != nil {
		log.Errorf("task ContractRecord Copy err:%+v", err)
	}
	return
}
