package handler

import (
	"github.com/gin-gonic/gin"
)

// Handler Handler
type Handler interface {
	Register(*gin.Engine)
}

var handlers = []Handler{
	NewHQHandler(),
}

// Register 注册所有的API入口
func Register(e *gin.Engine) {
	e.Use(ParseFormMiddleware)
	for _, h := range handlers {
		h.Register(e)
	}
}
