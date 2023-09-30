package model

import "time"

// Position 持仓表
type Position struct {
	ID           int64     `gorm:"column:id"`            // 主键ID
	UID          int64     `gorm:"column:uid"`           // 用户ID
	ContractID   int64     `gorm:"column:contract_id"`   // 合约编号
	EntrustID    int64     `gorm:"column:entrust_id"`    // 委托编号
	OrderTime    time.Time `gorm:"column:order_time"`    // 订单时间
	StockCode    string    `gorm:"column:stock_code"`    // 股票代码
	StockName    string    `gorm:"column:stock_name"`    // 股票名称
	Price        float64   `gorm:"column:price"`         // 持仓价格
	Amount       int64     `gorm:"column:amount"`        // 数量
	FreezeAmount int64     `gorm:"column:freeze_amount"` // 冻结股数
	Balance      float64   `gorm:"column:balance"`       // 成交金额
	CurPrice     float64   `gorm:"-"`                    // 当前现价
}

// CalculatePositionProfit 计算持仓盈亏
func CalculatePositionProfit(positions []*Position) float64 {
	profit := 0.00
	for _, p := range positions {
		profit += (p.CurPrice - p.Price) * float64(p.Amount)
	}
	return profit
}

// CalculatePositionMarketValue 持仓市值:amount*CurPrice
func CalculatePositionMarketValue(positions []*Position) float64 {
	marketValue := 0.00
	for _, p := range positions {
		marketValue += float64(p.Amount) * p.CurPrice
	}
	return marketValue
}

// CalculatePositionAsset 持仓资产:amount*price
func CalculatePositionAsset(positions []*Position) float64 {
	asset := 0.00
	for _, p := range positions {
		asset += float64(p.Amount) * p.Price
	}
	return asset
}
