package model

import (
	"time"
)

// BrokerEntrust 券商委托表
type BrokerEntrust struct {
	ID               int64     `gorm:"column:id"`                 // 主键ID
	UID              int64     `gorm:"column:uid"`                // 用户ID
	ContractID       int64     `gorm:"column:contract_id"`        // 合约编号
	BrokerID         int64     `gorm:"column:broker_id"`          // 券商ID
	EntrustID        int64     `gorm:"column:entrust_id"`         // 委托表ID
	OrderTime        time.Time `gorm:"column:order_time"`         // 订单时间
	StockCode        string    `gorm:"column:stock_code"`         // 股票代码
	StockName        string    `gorm:"column:stock_name"`         // 股票名称
	EntrustAmount    int64     `gorm:"column:entrust_amount"`     // 委托总股数
	EntrustPrice     float64   `gorm:"column:entrust_price"`      // 委托价格
	EntrustBalance   float64   `gorm:"column:entrust_balance"`    // 委托总金额
	DealAmount       int64     `gorm:"column:deal_amount"`        // 成交数量
	DealPrice        float64   `gorm:"column:deal_price"`         // 成交价格
	DealBalance      float64   `gorm:"column:deal_balance"`       // 成交总金额
	Status           int64     `gorm:"column:status"`             // 订单状态：1未成交 2成交 3已撤单
	EntrustBs        int64     `gorm:"column:entrust_bs"`         // 交易类型:1买入 2卖出
	EntrustProp      int64     `gorm:"column:entrust_prop"`       // 委托类型:1限价 2市价
	Fee              float64   `gorm:"column:fee"`                // 券商交易总手续费
	BrokerEntrustNo  string    `gorm:"column:broker_entrust_no"`  // 券商委托编号
	BrokerWithdrawNo string    `gorm:"column:broker_withdraw_no"` // 券商撤单编号
	Broker           *Broker   `gorm:"-"`                         // 券商信息
}

// IsFinallyState 是否终态
func (e *BrokerEntrust) IsFinallyState() bool {
	switch e.Status {
	case EntrustStatusTypeDeal, EntrustStatusTypeWithdraw, EntrustStatusTypePartDealPartWithdraw, EntrustStatusTypeCancel:
		return true
	}
	return false
}
