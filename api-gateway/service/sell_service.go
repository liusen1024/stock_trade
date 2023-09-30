package service

import (
	"context"
	"fmt"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"
	"sync"
)

// SellService 卖出服务
type SellService struct {
}

var (
	sellService *SellService
	sellOnce    sync.Once
)

// SellServiceInstance SellServiceInstance实例
func SellServiceInstance() *SellService {
	sellOnce.Do(func() {
		sellService = &SellService{}
	})
	return sellService
}

// CreateOrder 卖出订单成交
// 核心参数:deal_amount|status :deal_amount表示成交数量,status:终态
func (s *SellService) CreateOrder(ctx context.Context, entrust *model.Entrust) error {
	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()

	contract, err := dao.ContractDaoInstance().GetContractByIDWithTx(tx, entrust.ContractID)
	if err != nil {
		log.Errorf("GetContractByIDWithTx err:%+v", err)
		return err
	}

	position, err := dao.PositionDaoInstance().GetContractPositionByCodeWithTx(tx, entrust.ContractID, entrust.StockCode)
	if err != nil {
		log.Errorf("GetContractPositionByCode err:%+v", err)
		return serr.ErrBusiness("委托卖出失败:非持仓股票")
	}

	// 填写卖出记录
	sell, err := dao.SellDaoInstance().CreateWithTx(tx, &model.Sell{
		EntrustID:     entrust.ID,                                                     // 委托表ID
		UID:           entrust.UID,                                                    // 用户ID
		ContractID:    entrust.ContractID,                                             // 合约编号
		OrderTime:     entrust.OrderTime,                                              // 订单时间
		StockCode:     entrust.StockCode,                                              // 股票代码
		StockName:     entrust.StockName,                                              // 股票名称
		Price:         entrust.Price,                                                  // 价格
		Amount:        entrust.DealAmount,                                             // 数量
		Balance:       entrust.Price * float64(entrust.DealAmount),                    // 成交金额
		PositionPrice: position.Price,                                                 // 持仓价格
		Profit:        (entrust.Price - position.Price) * float64(entrust.DealAmount), // 盈亏金额
		EntrustProp:   entrust.EntrustProp,                                            // 委托类型:1限价 2市价
		Fee:           entrust.Fee,                                                    // 交易手续费
		PositionID:    position.ID,                                                    // 持仓表序号
		Mode:          entrust.Mode,                                                   // 类型:1 主动卖出 2系统平仓
		Reason:        entrust.Reason,                                                 // 系统平仓原因
	})
	if err != nil {
		log.Errorf("CreateWithTx err:%+v", err)
		return err
	}
	log.Infof("1.委托编号:%+v [sell]创建卖出记录成功:%+v", entrust.ID, sell)

	// 更新合约:盈亏
	contract.Money = contract.Money + sell.Profit
	if err := dao.ContractDaoInstance().UpdateWithTx(tx, contract); err != nil {
		log.Errorf("UpdateWithTx err:%+v", err)
		return err
	}
	log.Infof("2.委托编号:%+v [contract]合约盈亏金额:%+v,合约保证金:%+v", entrust.ID, sell.Profit, contract.Money)

	// 修改持仓股数
	if position.Amount == entrust.DealAmount {
		// 全部卖出
		if err := dao.PositionDaoInstance().DeleteWithTx(tx, position); err != nil {
			log.Errorf("DeleteWithTx err:%+v", err)
			return serr.ErrBusiness("卖出失败")
		}
	} else {
		// 非全仓卖出

		position.Amount = position.Amount - entrust.DealAmount
		position.FreezeAmount = position.FreezeAmount - entrust.Amount
		position.Balance = position.Price * float64(position.Amount)
		if err := dao.PositionDaoInstance().UpdateWithTx(tx, position); err != nil {
			log.Errorf("非全仓卖出失败:%+v", err)
			return serr.New(serr.ErrCodeBusinessFail, "委托交易失败")
		}
	}
	log.Infof("3.委托编号:%+v 更新持仓成功", entrust.ID)

	// 委托数量=卖出数量 || 委托状态等于终态
	// 1. 扣除卖出手续费
	contract.Money = contract.Money - entrust.Fee
	if err := dao.ContractDaoInstance().UpdateWithTx(tx, contract); err != nil {
		log.Errorf("UpdateWithTx err:%+v", err)
		return err
	}
	log.Infof("4.委托编号:%+v [contract]扣除卖出交易手续费:%+v", entrust.ID, entrust.Fee)

	// 2. 卖出手续费写入contract_fee表 & 盈亏填写contract_fee
	contractFee := &model.ContractFee{
		UID:        entrust.UID,
		ContractID: entrust.ContractID,
		Code:       entrust.StockCode,
		Name:       entrust.StockName,
		Amount:     entrust.DealAmount,
		OrderTime:  entrust.OrderTime,
		Direction:  model.ContractFeeDirectionPay,                  // 方向:1支出 2:收入
		Money:      entrust.Fee,                                    // 金额
		Detail:     fmt.Sprintf("卖出交易成功,扣取手续费:%0.2f", entrust.Fee), // 明细
		Type:       model.ContractFeeTypeSell,                      // 费用类型1:买入手续费 2:卖出手续费 3:合约利息 4:卖出盈亏 5:追加保证金 6:扩大资金 7:合约结算
	}
	if err := dao.ContractFeeDaoInstance().CreateWithTx(tx, contractFee); err != nil {
		log.Errorf("CreateWithTx err:%+v", err)
		return err
	}
	log.Infof("5.委托编号:%+v [contract_fee]创建卖出手续费:%+v", entrust.ID, contractFee)

	if err := dao.ContractFeeDaoInstance().CreateWithTx(tx, &model.ContractFee{
		UID:        entrust.UID,
		ContractID: entrust.ContractID,
		Code:       entrust.StockCode,
		Name:       entrust.StockName,
		Amount:     entrust.DealAmount,
		OrderTime:  entrust.OrderTime,
		Direction:  model.ContractFeeDirectionIncome,              // 方向:1支出 2:收入
		Money:      sell.Profit,                                   // 金额
		Detail:     fmt.Sprintf("卖出交易成功,盈亏金额:%0.2f", sell.Profit), // 明细
		Type:       model.ContractFeeTypeProfit,                   // 费用类型1:买入手续费 2:卖出手续费 3:合约利息 4:卖出盈亏 5:追加保证金 6:扩大资金 7:合约结算
	}); err != nil {
		log.Errorf("CreateWithTx err:%+v", err)
		return err
	}

	// 3. 填写msg表(手续费+盈亏金额)
	if err := dao.MsgDaoInstance().CreateWithTx(tx, &model.Msg{
		UID:   entrust.UID,         // 用户ID
		Title: fmt.Sprintf("委托成交"), // 标题
		Content: fmt.Sprintf("合约[%d]:%s(%s)卖出成交!成交数量%d股,成交均价%0.2f元,成交金额%0.2f元,交易手续费%0.2f元,盈亏金额%0.2f元",
			entrust.ContractID, entrust.StockName, entrust.StockCode, sell.Amount, sell.Price, sell.Balance, sell.Fee, sell.Profit), // 内容
		CreateTime: entrust.OrderTime,
	}); err != nil {
		return err
	}

	if err := dao.EntrustDaoInstance().UpdateWithTx(tx, entrust); err != nil {
		log.Errorf("更新委托表失败:%+v", err)
		return err
	}

	// 同步entrust表状态到brokerEntrust
	if entrust.IsBrokerEntrust && len(entrust.BrokerEntrust) > 0 {
		if err := dao.BrokerEntrustDaoInstance().MCreateWithTx(tx, entrust.BrokerEntrust); err != nil {
			log.Errorf("更新券商委托表失败:%+v", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Errorf("委托编号:%+v,提交失败:%+v", entrust.ID, err)
		return err
	}
	log.Infof("委托编号:%+v,交易成功!", entrust.ID)

	// 更新可用资金
	if err := ContractServiceInstance().UpdateValMoneyByID(ctx, entrust.ContractID); err != nil {
		log.Errorf("刷新资金失败:%+v", err)
	}
	return nil
}
