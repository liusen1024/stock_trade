package service

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/common/log"
	"sync"
	"time"
)

// BuyService 买入服务
type BuyService struct {
}

var (
	buyService *BuyService
	buyOnce    sync.Once
)

// BuyServiceInstance BuyServiceInstance实例
func BuyServiceInstance() *BuyService {
	buyOnce.Do(func() {
		buyService = &BuyService{}
	})
	return buyService
}

// CreateOrder 买入订单成交
// 订单终态：entrust.Amount等于=entrust.DealAmount 或者 参数entrust.status为终态时
func (s *BuyService) CreateOrder(ctx context.Context, entrust *model.Entrust) error {
	tx := db.StockDB().WithContext(ctx).Begin()
	defer tx.Rollback()

	contract, err := dao.ContractDaoInstance().GetContractByIDWithTx(tx, entrust.ContractID)
	if err != nil {
		log.Errorf("GetContractByIDWithTx err:%+v", err)
		return err
	}

	position, err := dao.PositionDaoInstance().GetContractPositionByCodeWithTx(tx, entrust.ContractID, entrust.StockCode)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("GetPositionByCode err:%+v", err)
		return err
	}

	if position == nil {
		// 无持仓,新建持仓
		p, err := dao.PositionDaoInstance().CreateWithTx(tx, &model.Position{
			UID:          entrust.UID,                                 // 用户ID
			ContractID:   entrust.ContractID,                          // 合约编号
			EntrustID:    entrust.ID,                                  // 委托编号
			OrderTime:    entrust.OrderTime,                           // 订单时间
			StockCode:    entrust.StockCode,                           // 股票代码
			StockName:    entrust.StockName,                           // 股票名称
			Price:        entrust.Price,                               // 持仓价格
			Amount:       entrust.DealAmount,                          // 数量
			Balance:      entrust.Price * float64(entrust.DealAmount), // 成交金额
			FreezeAmount: entrust.DealAmount,                          // 冻结股数
		})
		if err != nil {
			log.Errorf("创建持仓表失败:%+v", err)
			return err
		}
		position = p
		log.Infof("1.委托编号:%+v [position]新建持仓成功:%+v", entrust.ID, position)
	} else {
		// 非第一次买入则更新持仓记录 : 股数,num=num+%s,freezenum=freezenum+%s,price=%s
		position.Price = (position.Price*float64(position.Amount) + entrust.Price*float64(entrust.DealAmount)) / float64(position.Amount+entrust.DealAmount)
		position.Amount = position.Amount + entrust.DealAmount
		position.Balance = position.Price * float64(position.Amount)
		position.FreezeAmount = position.FreezeAmount + entrust.DealAmount
		if err := dao.PositionDaoInstance().UpdateWithTx(tx, position); err != nil {
			log.Errorf("交易错误:更新持仓表错误:%+v", err)
			return err
		}
		log.Infof("1.委托编号:%+v [position]更新持仓成功:%+v", entrust.ID, position)
	}

	// 买入记录
	buy := &model.Buy{
		EntrustID:   entrust.ID,
		UID:         entrust.UID,
		ContractID:  entrust.ContractID,
		OrderTime:   time.Now(),
		StockCode:   entrust.StockCode,
		StockName:   entrust.StockName,
		Price:       entrust.Price,
		Amount:      entrust.DealAmount,
		Balance:     entrust.Price * float64(entrust.DealAmount),
		EntrustProp: entrust.EntrustProp,
		Fee:         entrust.Fee,
		PositionID:  position.ID}
	if err := dao.BuyDaoInstance().CreateWithTx(tx, buy); err != nil {
		log.Errorf("CreateWithTx err:%+v", err)
		return err
	}
	log.Infof("2.委托编号:%+v [buy]创建买入记录成功:%+v", entrust.ID, buy)

	// 1. 扣除买入手续费
	contract.Money -= entrust.Fee
	if err := dao.ContractDaoInstance().UpdateWithTx(tx, contract); err != nil {
		log.Errorf("contract err:%+v", err)
		return err
	}
	log.Infof("3.委托编号:%+v [contract]扣除手续费:%+v 成功", entrust.ID, entrust.Fee)

	// 2. 写入contract_fee表
	contractFee := &model.ContractFee{
		UID:        entrust.UID,
		ContractID: entrust.ContractID,
		Code:       entrust.StockCode,                              // 股票代码
		Name:       entrust.StockName,                              // 股票名称
		Amount:     entrust.DealAmount,                             // 股票交易数量
		OrderTime:  entrust.OrderTime,                              // 订单时间
		Direction:  model.ContractFeeDirectionPay,                  // 方向:1支出 2:收入
		Money:      entrust.Fee,                                    // 金额
		Detail:     fmt.Sprintf("买入交易成功,扣取手续费:%0.2f", entrust.Fee), // 明细
		Type:       model.ContractFeeTypeBuy,                       // 费用类型1:买入手续费 2:卖出手续费 3:合约利息 4:卖出盈亏 5:追加保证金 6:扩大资金 7:合约结算
	}
	if err := dao.ContractFeeDaoInstance().CreateWithTx(tx, contractFee); err != nil {
		log.Errorf("contract_fee err:%+v", err)
		return err
	}
	log.Infof("4.委托编号:%+v [contract_fee]创建合约费用记录:%+v 成功", entrust.ID, contractFee)

	// 3. 填写msg表
	msg := &model.Msg{
		UID:   entrust.UID,         // 用户ID
		Title: fmt.Sprintf("委托成交"), // 标题
		Content: fmt.Sprintf("合约[%d]:%s(%s)买入成交!成交数量%d股，成交均价%0.2f元,成交金额%0.2f元,交易手续费%0.2f元",
			entrust.ContractID, entrust.StockName, entrust.StockCode, entrust.DealAmount, entrust.Price, float64(entrust.DealAmount)*entrust.Price, entrust.Fee), // 内容
		CreateTime: entrust.OrderTime,
	}
	if err := dao.MsgDaoInstance().CreateWithTx(tx, msg); err != nil {
		log.Errorf("msg err:%+v", err)
		return err
	}
	log.Infof("5.委托编号:%+v [msg]创建消息:%+v 成功", entrust.ID, msg)

	entrust.PositionID = position.ID
	if err := dao.EntrustDaoInstance().UpdateWithTx(tx, entrust); err != nil {
		log.Errorf("create entrust err:%+v", err)
		return err
	}
	log.Infof("6.委托编号:%+v [entrust]更新委托表:%+v 成功", entrust.ID, entrust)

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
