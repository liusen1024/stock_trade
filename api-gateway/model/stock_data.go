package model

import "time"

// /////////////////////////////////StockData股票数据表///////////////////////////////////
const (
	StockDataStatusEnable  = 1
	StockDataStatusDisable = 2
)

// XueQiuHotStock 雪球热门股票
type XueQiuHotStock struct {
	Code string
	Name string
}

// StockData 股票数据
type StockData struct {
	ID     int64     `gorm:"column:id"`
	Code   string    `gorm:"column:code"`    // 股票代码
	Name   string    `gorm:"column:name"`    // 股票名称
	IPODay time.Time `gorm:"column:ipo_day"` // IPO日期
	Status int64     `gorm:"column:status"`  // 交易状态:1允许交易 2不允许交易

	IsMargin  bool  `gorm:"-"` // 融资融券股票
	HxSignal  int64 `gorm:"-"` // 华兴操盘线信号 0无信号,1:买入,2:卖出
	XgbSignal int   `gorm:"-"` // 选股宝技术面分析信号 0无信号 1:走势良好 2:走势很弱

	// 选股宝技术面分析信号
	// 同花顺信号
	// 富途信号
	// 雪球信号

}

///////////////////////////////////StockData股票数据表///////////////////////////////////
