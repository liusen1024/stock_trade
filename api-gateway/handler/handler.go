package handler

import "github.com/gin-gonic/gin"

// Handler Handler
type Handler interface {
	Register(*gin.Engine)
}

var handlers = []Handler{
	NewUserHandler(),      // 用户
	NewSmsHandler(),       // 短信发送
	NewHomeHandler(),      // 首页
	NewPortfolioHandler(), // 自选股
	NewMarketHandler(),    // 市场
	NewContractHandler(),  // 合约
	NewTradeHandler(),     // 交易
	NewSearchHandler(),    // 搜索
	NewMyHandler(),        // 我的
	NewStockHandler(),     // 股票列表

}

// Register 注册所有的API入口
func Register(e *gin.Engine) {
	for _, h := range handlers {
		h.Register(e)
	}
}
