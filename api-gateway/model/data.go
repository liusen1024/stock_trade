package model

import (
	"sort"
	"time"
)

// Data 存储各种数据
type Data struct {
	UpdateTime         time.Time           // 数据更新时间
	MarketOverView     *MarketOverView     // 市场概况
	Rank               *Rank               // 市场排名
	XGBMarketIndicator *XGBMarketIndicator // 选股宝市场指标
	HotStock           []*HotStock         // 雪球热门股票
	IndustrySector     []*Sector           // 行业版块
	ConceptSector      []*Sector           // 概念版块
	Index              []*Index            // 上证指数|深证指数|创业板指数
}

// Index 指数
type Index struct {
	Code       string  `json:"code"`        // 股票代码
	Name       string  `json:"name"`        // 股票名称
	Price      float64 `json:"price"`       // 最新价
	Chg        float64 `json:"chg"`         // 涨跌额
	ChgPercent float64 `json:"chg_percent"` // 涨跌幅度
}

// Sector 板块
type Sector struct {
	Code            string  `json:"code"`              // 板块代码
	Name            string  `json:"name"`              // 板块名称
	ChgPercent      float64 `json:"chg_percent"`       // 板块涨跌幅度
	UpAmount        int64   `json:"up_amount"`         // 上涨家数
	DownAmount      int64   `json:"down_amount"`       // 下跌家数
	Turnover        float64 `json:"turnover"`          // 版块换手率
	StockCode       string  `json:"stock_code"`        // 股票代码
	StockName       string  `json:"stock_name"`        // 股票名称
	StockPrice      float64 `json:"stock_price"`       // 股票价格
	StockChgPercent float64 `json:"stock_chg_percent"` // 股票涨跌幅度
}

// XGBMarketIndicator 选股宝市场指标
type XGBMarketIndicator struct {
	LimitDownCount         int64   `json:"limit_down_count"`           // 跌停
	LimitUpBrokenCount     int64   `json:"limit_up_broken_count"`      // 炸板数量
	LimitUpBrokenRatio     float64 `json:"limit_up_broken_ratio"`      // 炸板率
	LimitUpCount           int64   `json:"limit_up_count"`             // 涨停
	YesterdayLimitUpAvgPcp float64 `json:"yesterday_limit_up_avg_pcp"` // 昨涨停今表现
}

// MarketOverView 市场概况
type MarketOverView struct {
	UpAmount     int64 // 上涨家数
	DownAmount   int64 // 下跌家数
	FlatAmount   int64 // 平家数量
	OverViewItem []*OverViewItem
}

// OverViewItem 市场概况分布
type OverViewItem struct {
	Amount int64  `json:"amount"` // 数量
	Desc   string `json:"desc"`   // 描述: 涨停, > 7
}

type Rank struct {
	UpRank           []*StockItem `json:"up_rank"`             // 涨幅榜
	DownRank         []*StockItem `json:"down_rank"`           // 跌幅榜
	TurnoverRank     []*StockItem `json:"turnover_rank"`       // 换手榜
	VolumeRank       []*StockItem `json:"volume_rank"`         // 成交榜
	NetInBalanceRank []*StockItem `json:"net_in_balance_rank"` // 净流入榜
}

// StockItem 股票
type StockItem struct {
	Code       string  `json:"code"`        // 股票代码
	Name       string  `json:"name"`        // 股票名称
	Price      float64 `json:"price"`       // 最新价格
	ChgPercent float64 `json:"chg_percent"` // 涨跌幅
}

func Process(list []*StockItem, sortBy, orderBy string) []*StockItem {
	if len(sortBy) == 0 {
		return list
	}
	// 根据价格排序
	switch sortBy {
	case "price": // 根据价格排序
		if orderBy == "asc" {
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].Price < list[j].Price
			})
		} else if orderBy == "desc" {
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].Price > list[j].Price
			})
		}

	case "chg_percent": // 根据涨跌幅排序
		if orderBy == "asc" {
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].ChgPercent < list[j].ChgPercent
			})
		} else if orderBy == "desc" {
			sort.SliceStable(list, func(i, j int) bool {
				return list[i].ChgPercent > list[j].ChgPercent
			})
		}
	}
	return list
}
