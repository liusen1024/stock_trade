package model

// Home 首页
type Home struct {
	UpAmount               int64        `json:"up_amount"`                  // 上涨家数
	FlatAmount             int64        `json:"flat_amount"`                // 平家数
	DownAmount             int64        `json:"down_amount"`                // 下跌家数
	MarketTemperature      float64      `json:"market_temperature"`         // 市场热度
	LimitUpCount           int64        `json:"limit_up_count"`             // 涨停
	LimitDownCount         int64        `json:"limit_down_count"`           // 跌停
	LimitUpBrokenRatio     float64      `json:"limit_up_broken_ratio"`      // 炸板率
	LimitUpBrokenCount     int64        `json:"limit_up_broken_count"`      // 炸板数量
	YesterdayLimitUpAvgPcp float64      `json:"yesterday_limit_up_avg_pcp"` // 昨涨停今表现
	HotStock               []*HotStock  `json:"hot_stock"`                  // 热门股票
	UpRank                 []*StockItem `json:"up_rank"`                    // 涨幅榜
	DownRank               []*StockItem `json:"down_rank"`                  // 跌幅榜
	TurnoverRank           []*StockItem `json:"turnover_rank"`              // 换手榜
	VolumeRank             []*StockItem `json:"volume_rank"`                // 成交榜
	NetInBalanceRank       []*StockItem `json:"net_in_balance_rank"`        // 净流入榜
}

// HotStock 热门股票
type HotStock struct {
	Code    string  `json:"code"`    // 股票代码
	Name    string  `json:"name"`    // 股票名称
	Price   float64 `json:"price"`   // 股票价格
	Percent float64 `json:"percent"` // 涨跌幅
	Chg     float64 `json:"chg"`     // 涨跌额
}

// Pool 股票池子
type Pool struct {
	PoolItem []*PoolItem `json:"items"`
	Size     int64       `json:"size"`
}

// PoolItem 股票
type PoolItem struct {
	Code    string  `json:"code"`    // 股票代码
	Name    string  `json:"name"`    // 股票名称
	Price   float64 `json:"price"`   // 股票价格
	Percent float64 `json:"percent"` // 涨跌幅
	Extra   string  `json:"extra"`   // 额外信息
}
