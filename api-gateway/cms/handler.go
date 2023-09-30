package handler

import (
	"github.com/gin-gonic/gin"
)

// Handler Handler
type Handler interface {
	Register(*gin.Engine)
}

var handlers = []Handler{
	NewCMSHandler(),
	NewUserHandler(),
	NewAgentHandler(),
	NewTradeHandler(),
	NewContractHandler(),
	NewBrokerHandler(),
	NewStockHandler(),
	NewSystemHandler(),
	NewLogHandler(),
}

// Register 注册所有的API入口
func Register(e *gin.Engine) {
	e.Use(Auth) // session 鉴权
	e.Use(ACL)  // post,get 鉴权
	e.Use(ParseFormMiddleware)
	for _, h := range handlers {
		h.Register(e)
	}
}
