package util

import (
	"strconv"
	"strings"
)

const (
	StockTypeNormal  = 0 // 普通股票
	StockTypeKCBBORD = 1 // 科创板
	StockTypeCYBBORD = 2 // 创业板
	StockTypeBJ      = 3 // 北交所股票
)

// StockBord 股票所属板块
func StockBord(code string) int64 {
	if strings.HasPrefix(code, "688") {
		return StockTypeKCBBORD
	} else if strings.HasPrefix(code, "300") {
		return StockTypeCYBBORD
	} else if strings.HasPrefix(code, "83") || strings.HasPrefix(code, "87") || strings.HasPrefix(code, "88") {
		return StockTypeBJ
	}
	return 0
}

const (
	StockMarketTypeSZ = "SZ"
	StockMarketTypeSH = "SH"
	StockMarketTypeBJ = "BJ"
)

// GetStockMarketType 获取股票所属证券市场
func GetStockMarketType(stockCode string) string {
	code, err := strconv.ParseInt(stockCode, 10, 64)
	if err != nil {
		return ""
	}
	if code < 400000 {
		return StockMarketTypeSZ
	}
	if code >= 400000 && code < 500000 {
		return StockMarketTypeBJ
	}
	if code >= 600000 && code < 800000 {
		return StockMarketTypeSH
	}
	// 800000 是北交所股票
	return StockMarketTypeBJ
}
