package handler

import (
	"stock/api-gateway/service"
	"stock/api-gateway/util"

	"github.com/gin-gonic/gin"
)

// HomeHandler 首页服务handler
type HomeHandler struct {
}

// NewHomeHandler 单例
func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

// Register 注册handler
func (h *HomeHandler) Register(e *gin.Engine) {
	// 首页-市场热度
	e.GET("/home/get", JSONWrapper(h.GetHome))
	// 首页-全部热股
	e.GET("/home/all_hot_stock", JSONWrapper(h.AllHotStock))
}

// GetHome 首页
func (h *HomeHandler) GetHome(c *gin.Context) (interface{}, error) {
	return service.HomeServiceInstance().GetHome(util.RPCContext(c))
}

func (h *HomeHandler) AllHotStock(c *gin.Context) (interface{}, error) {
	return service.HomeServiceInstance().AllHotStock(util.RPCContext(c))
}
