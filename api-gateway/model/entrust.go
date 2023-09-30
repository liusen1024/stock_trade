package model

import (
	"fmt"
	"time"
)

///////////////////////////////////entrust自选股表///////////////////////////////////

const (
	EntrustBsTypeBuy  = 1 // 委托类型:买入
	EntrustBsTypeSell = 2 // 委托类型:卖出

	EntrustStatusTypeUnDeal               = 1 // 委托状态:未成交
	EntrustStatusTypeDeal                 = 2 // 委托状态:全部成交
	EntrustStatusTypeWithdraw             = 3 // 委托状态:已撤单
	EntrustStatusTypePartDealPartWithdraw = 4 // 委托状态:部成部撤
	EntrustStatusTypeWithdrawing          = 5 // 委托状态:等待撤单(用户发起撤单后的状态)
	EntrustStatusTypeReported             = 6 // 委托状态:已申报,未成交
	EntrustStatusTypePartDeal             = 7 // 部分成交
	EntrustStatusTypeCancel               = 8 // 委托状态:废单

	EntrustPropTypeLimitPrice  = 1 // 限价
	EntrustPropTypeMarketPrice = 2 // 市价
)

var EntrustStatusMap = map[int64]string{
	EntrustStatusTypeUnDeal:               "未成交",
	EntrustStatusTypeDeal:                 "全部成交",
	EntrustStatusTypeWithdraw:             "已撤单",
	EntrustStatusTypePartDealPartWithdraw: "部成部撤",
	EntrustStatusTypeWithdrawing:          "等待撤单",
	EntrustStatusTypeReported:             "已申报",
	EntrustStatusTypePartDeal:             "部分成交",
	EntrustStatusTypeCancel:               "废单",
}

// Entrust 委托表
type Entrust struct {
	ID              int64            `gorm:"column:id"`                // 主键ID
	UID             int64            `gorm:"column:uid"`               // 用户ID
	ContractID      int64            `gorm:"column:contract_id"`       // 合约编号
	OrderTime       time.Time        `gorm:"column:order_time"`        // 订单时间
	StockCode       string           `gorm:"column:stock_code"`        // 股票代码
	StockName       string           `gorm:"column:stock_name"`        // 股票名称
	Amount          int64            `gorm:"column:amount"`            // 数量(股)
	Price           float64          `gorm:"column:price"`             // 价格
	Balance         float64          `gorm:"column:balance"`           // 委托金额
	DealAmount      int64            `gorm:"column:deal_amount"`       // 数量(股)
	Status          int64            `gorm:"column:status"`            // 委托状态:1未成交 2成交 3已撤单 4部成部撤 5等待撤单(用户发起撤单后的状态) 6已申报,未成交 7部分成交 8废单
	EntrustBS       int64            `gorm:"column:entrust_bs"`        // 交易类型:1买入 2卖出
	EntrustProp     int64            `gorm:"column:entrust_prop"`      // 委托类型:1限价 2市价
	PositionID      int64            `gorm:"column:position_id"`       // 持仓表id(卖出时需填写)
	Fee             float64          `gorm:"column:fee"`               // 总交易费用
	IsBrokerEntrust bool             `gorm:"column:is_broker_entrust"` // 是否券商委托
	Remark          string           `gorm:"column:remark"`            // 备注:委托失败
	Mode            int64            `gorm:"column:mode"`              // 类型:0 主动卖出 1系统平仓
	Reason          string           `gorm:"-"`                        // 系统平仓原因
	BrokerEntrust   []*BrokerEntrust `gorm:"-"`                        // 券商委托
}

func (i *Entrust) ConvertEntrustBsToString() string {
	if i.EntrustBS == EntrustBsTypeBuy {
		return "买入"
	}
	return "卖出"
}

// CancelBrokerEntrust 取消券商委托
func (e *Entrust) CancelBrokerEntrust() {
	e.IsBrokerEntrust = false
}

func EntrustLockKey(entrustID int64) string {
	return fmt.Sprintf("redis_lock_entrust_id_%d", entrustID)
}

// IsFinallyState 是否终态
func (e *Entrust) IsFinallyState() bool {
	switch e.Status {
	case EntrustStatusTypeDeal, EntrustStatusTypeWithdraw, EntrustStatusTypePartDealPartWithdraw, EntrustStatusTypeCancel:
		return true
	}
	return false
}
