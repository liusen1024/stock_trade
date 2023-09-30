package model

import "time"

// 合约类型
const (
	DividendTypeStockConversion              = 1 // 送股
	DividendTypeBonusShare                   = 2 // 现金分红
	DividendTypeBonusShareAndStockConversion = 3 // 现金分红+送股
)

type Dividend struct {
	ID             int64     `gorm:"column:id"`              // 委托表ID
	UID            int64     `gorm:"column:uid"`             // 用户ID
	ContractID     int64     `gorm:"column:contract_id"`     // 合约编号
	PositionID     int64     `gorm:"column:position_id"`     // 持仓编号
	OrderTime      time.Time `gorm:"column:order_time"`      // 订单时间
	StockCode      string    `gorm:"column:stock_code"`      // 股票代码
	StockName      string    `gorm:"column:stock_name"`      // 股票名称
	PositionPrice  float64   `gorm:"column:position_price"`  // 持仓价格
	PositionAmount int64     `gorm:"column:position_amount"` // 持仓股数
	IsBuyBack      bool      `gorm:"column:is_buy_back"`     // 是否零股回购
	BuyBackAmount  int64     `gorm:"column:buy_back_amount"` // 零股回购数量
	BuyBackPrice   float64   `gorm:"column:buy_back_price"`  // 零股回购价格
	DividendMoney  float64   `gorm:"column:dividend_money"`  // 现金分红金额
	DividendAmount int64     `gorm:"column:dividend_amount"` // 转股数量
	Type           int64     `gorm:"column:type"`            // 类型:1分红送股 2:现金分红
	PlanExplain    string    `gorm:"column:plan_explain"`    // 方案说明
}
