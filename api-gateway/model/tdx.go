package model

import (
	"stock/api-gateway/util"
	"strconv"
	"strings"
	"time"
)

// TDXBrokerFund 通达信券商资金
type TDXBrokerFund struct {
	ClientID    int64   // clientID
	FundAccount string  // 资金账号
	ValMoney    float64 // 可用资金
	Asset       float64 // 总资产
	MarketValue float64 // 市值
}

// TDXTodayEntrust 通达信今日委托
type TDXTodayEntrust struct {
	ClientID    int64     // 客户ID
	FundAccount string    // 资金账号
	EntrustTime time.Time // 委托时间
	EntrustNo   string    // 委托编号
	StockCode   string    // 股票代码
	StockName   string    // 股票名称
	EntrustBs   int       // 委托类型:买入,卖出
	Price       float64   // 委托价格
	Amount      int64     // 委托数量
	DealAmount  int64     // 成交数量
	DealBalance float64   // 成交金额
	DealPrice   float64   // 成交价格
	Status      string    // 状态:已报,已成,已撤,部撤
}

// TDXPosition 通达信券商持仓
type TDXPosition struct {
	ClientID      int64   // 券商客户ID
	FundAccount   string  // 资金账号
	StockCode     string  // 股票代码
	StockName     string  // 股票名称
	Amount        int64   // 持仓数量
	FreezeAmount  int64   // 冻结数量
	PositionPrice float64 `json:"position_price"` // 持仓价格
	CurrentPrice  float64 `json:"current_price"`  // 当前价格
}

// TDXWithdraw 通达信券商可撤单列表
type TDXWithdraw struct {
	ClientID      int64     // 券商客户ID
	FundAccount   string    // 资金账号
	EntrustTime   time.Time // 委托时间
	StockCode     string    // 股票代码
	StockName     string    // 股票名称
	EntrustBs     int       // 委托类型:买入,卖出
	EntrustPrice  float64   // 委托价格
	EntrustAmount int64     // 委托数量
	EntrustNo     string    // 委托编号
	DealAmount    int64     // 成交数量
}

// ParseTdxTodayEntrust 今日委托查询
func ParseTdxTodayEntrust(broker *Broker, src [][]string) []*TDXTodayEntrust {
	list := make([]*TDXTodayEntrust, 0)
	columnMap := make(map[int]string) // map<栏目序号>栏目名
	for col, items := range src {
		if col == 0 {
			for index, column := range items {
				columnMap[index] = column
			}
			continue
		}
		entrust := &TDXTodayEntrust{
			ClientID:    broker.ClientID,
			FundAccount: broker.FundAccount,
		}
		for index, v := range items {
			column, ok := columnMap[index]
			if !ok {
				continue
			}
			switch column {
			case "委托时间":
				entrust.EntrustTime, _ = time.Parse("15:04:05", v)
			case "证券代码":
				entrust.StockCode = v
			case "证券名称":
				entrust.StockName = v
			case "买卖标志":
				{
					entrust.EntrustBs = EntrustBsTypeSell
					if v == "0" || strings.Contains(v, "买") {
						entrust.EntrustBs = EntrustBsTypeBuy
					}
				}
			case "委托价格":
				entrust.Price, _ = strconv.ParseFloat(v, 10)
			case "委托数量":
				entrust.Amount, _ = strconv.ParseInt(v, 10, 64)
			case "成交价格":
				entrust.DealPrice, _ = strconv.ParseFloat(v, 10)
			case "成交数量":
				entrust.DealAmount, _ = strconv.ParseInt(v, 10, 64)
			case "成交金额":
				entrust.DealBalance, _ = strconv.ParseFloat(v, 10)
			case "状态说明":
				entrust.Status = v
			case "委托编号":
				entrust.EntrustNo = v
			}
		}

		entrust.DealBalance = util.FloatRound(entrust.DealPrice*float64(entrust.DealAmount), 2)
		if len(entrust.StockCode) > 0 {
			list = append(list, entrust)
		}
	}
	return list
}

//
//// parseSWHY 解析申万宏源
//func parseSWHY(broker *Broker, src [][]string) []*TDXTodayEntrust {
//	list := make([]*TDXTodayEntrust, 0)
//	for col, items := range src {
//		// 跳过表头
//		if col == 0 {
//			continue
//		}
//		entrust := &TDXTodayEntrust{
//			ClientID:    broker.ClientID,
//			FundAccount: broker.FundAccount,
//		}
//		for index, v := range items {
//			switch index {
//			case 0: // 委托时间
//				entrustTime, _ := time.Parse("15:04:05", v)
//				entrust.EntrustTime = entrustTime
//			case 1: // 委托编号
//				entrust.EntrustNo = v
//			case 3: // 证券代码
//				entrust.StockCode = v
//			case 4: // 证券名称
//				entrust.StockName = v
//			case 5: // 买卖标志
//				entrust.EntrustBs = EntrustBsTypeBuy
//				if v == "1" {
//					entrust.EntrustBs = EntrustBsTypeSell
//				}
//			case 8: // 委托价格
//				entrust.Price, _ = strconv.ParseFloat(v, 10)
//			case 9: // 委托数量
//				entrust.Amount, _ = strconv.ParseInt(v, 10, 64)
//			case 11: // 成交数量
//				entrust.DealAmount, _ = strconv.ParseInt(v, 10, 64)
//			case 12: // 成交金额
//				entrust.DealBalance, _ = strconv.ParseFloat(v, 10)
//			case 13: // 成交价格
//				entrust.DealPrice, _ = strconv.ParseFloat(v, 10)
//			case 17: // 状态说明
//				entrust.Status = v
//			}
//		}
//
//		if len(entrust.StockCode) != 0 {
//			list = append(list, entrust)
//		}
//	}
//	return list
//}
