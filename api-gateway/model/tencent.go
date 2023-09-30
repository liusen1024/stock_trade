package model

import "time"

const MaxURLLen = 2048

// TencentQuote 腾讯股票行情
type TencentQuote struct {
	Name             string    // 股票名称
	Code             string    // 股票代码
	CurrentPrice     float64   // 当前价格
	ClosePrice       float64   // 昨收
	OpenPrice        float64   // 今开
	TotalVol         int64     // 成交量(万手)
	BuyPrice1        float64   // 买一价
	BuyVol1          int64     // 买一量
	BuyPrice2        float64   // 买二价
	BuyVol2          int64     // 买二量
	BuyPrice3        float64   // 买三价
	BuyVol3          int64     // 买三量
	BuyPrice4        float64   // 买四价
	BuyVol4          int64     // 买四量
	BuyPrice5        float64   // 买五价
	BuyVol5          int64     // 买五量
	SellPrice1       float64   // 卖一价
	SellVol1         int64     // 卖一量
	SellPrice2       float64   // 卖二价
	SellVol2         int64     // 卖二量
	SellPrice3       float64   // 卖三价
	SellVol3         int64     // 卖三量
	SellPrice4       float64   // 卖四价
	SellVol4         int64     // 卖四量
	SellPrice5       float64   // 卖五价
	SellVol5         int64     // 卖五量
	Time             string    // 时间
	Chg              float64   // 涨跌额
	ChgPercent       float64   // 涨跌幅度
	HighPx           float64   // 最高价
	LowPx            float64   // 最低价
	TotalAmount      float64   // 37成交额(万)
	TurnOverRate     float64   // 38换手率
	Pe               float64   // 39市盈率
	FloatMarketValue float64   // 44流通市值
	TotalMarketValue float64   // 45总市值
	Pb               float64   // 46市净率
	LimitUpPrice     float64   // 47涨停价
	LimitDownPrice   float64   // 48跌停价
	DataTime         time.Time // 数据有效时间:从网络读取到数据生成的时间
}

// 定义盘口字段
const (
	Name             = 1  // 股票名称
	Code             = 2  // 股票代码
	CurrentPrice     = 3  // 当前价格
	ClosePrice       = 4  // 昨收
	OpenPrice        = 5  // 今开
	TotalVol         = 6  // 成交量(手)
	BuyPrice1        = 9  // 买一价
	BuyVol1          = 10 // 买一量
	BuyPrice2        = 11 // 买二价
	BuyVol2          = 12 // 买二量
	BuyPrice3        = 13 // 买三价
	BuyVol3          = 14 // 买三量
	BuyPrice4        = 15 // 买四价
	BuyVol4          = 16 // 买四量
	BuyPrice5        = 17 // 买五价
	BuyVol5          = 18 // 买五量
	SellPrice1       = 19 // 卖一价
	SellVol1         = 20 // 卖一量
	SellPrice2       = 21 // 卖二价
	SellVol2         = 22 // 卖二量
	SellPrice3       = 23 // 卖三价
	SellVol3         = 24 // 卖三量
	SellPrice4       = 25 // 卖四价
	SellVol4         = 26 // 卖四量
	SellPrice5       = 27 // 卖五价
	SellVol5         = 28 // 卖五量
	Time             = 30 // 时间
	Chg              = 31 // 涨跌
	ChgPercent       = 32 // 涨跌幅度
	HighPx           = 33 // 最高价
	LowPx            = 34 // 最低价
	TotalAmount      = 37 // 37成交额(万)
	TurnOverRate     = 38 // 38换手率
	Pe               = 39 // 39市盈率
	FloatMarketValue = 44 // 44流通市值
	TotalMarketValue = 45 // 45总市值
	Pb               = 46 // 46市净率
	LimitUpPrice     = 47 // 47涨停价
	LimitDownPrice   = 48 // 48跌停价
)
