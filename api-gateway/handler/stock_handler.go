package handler

import (
	"stock/api-gateway/service"
	"stock/api-gateway/util"

	"github.com/gin-gonic/gin"
)

// StockHandler handler
type StockHandler struct {
}

// NewStockHandler 单例
func NewStockHandler() *StockHandler {
	return &StockHandler{}
}

// Register 注册handler
func (h *StockHandler) Register(e *gin.Engine) {
	// 热门搜索股票
	e.GET("/stock/list", JSONWrapper(h.GetStockList))

}

// GetStockList 股票列表
func (h *StockHandler) GetStockList(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type stockItem struct {
		StockCode string `json:"stock_code"`
		StockName string `json:"stock_name"`
	}
	list := make([]*stockItem, 0)
	stocks, err := service.StockDataServiceInstance().GetStocks(ctx)
	if err != nil {
		return make([]*stock, 0), nil
	}
	for _, it := range stocks {
		list = append(list, &stockItem{
			StockCode: it.Code,
			StockName: it.Name,
		})
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}
