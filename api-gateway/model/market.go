package model

// Market 市场
type Market struct {
	Index             []*Index        `json:"index"`               // 指数
	OverViewItem      []*OverViewItem `json:"overview"`            // 市场概况
	UpAmount          int64           `json:"up_amount"`           // 上涨家数
	FlatAmount        int64           `json:"flat_amount"`         // 平家数
	DownAmount        int64           `json:"down_amount"`         // 下跌家数
	MarketTemperature float64         `json:"market_temperature"`  // 市场热度
	Industry          []*Sector       `json:"industry"`            // 行业版块
	Concept           []*Sector       `json:"concept"`             // 概念板块
	UpRank            []*StockItem    `json:"up_rank"`             // 涨幅榜
	DownRank          []*StockItem    `json:"down_rank"`           // 跌幅榜
	TurnoverRank      []*StockItem    `json:"turnover_rank"`       // 换手榜
	VolumeRank        []*StockItem    `json:"volume_rank"`         // 成交榜
	NetInBalanceRank  []*StockItem    `json:"net_in_balance_rank"` // 净流入榜
}
