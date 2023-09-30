package handler

import (
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/service"
	"stock/api-gateway/util"

	"github.com/gin-gonic/gin"
)

// PortfolioHandler 自选股handler
type PortfolioHandler struct {
}

// NewPortfolioHandler 单例
func NewPortfolioHandler() *PortfolioHandler {
	return &PortfolioHandler{}
}

// Register 注册handler
func (h *PortfolioHandler) Register(e *gin.Engine) {
	e.GET("/portfolio/list", JSONWrapper(h.List))
	e.GET("/portfolio/delete", JSONWrapper(h.Delete))
	e.GET("/portfolio/add", JSONWrapper(h.Add))
}

// List 自选股
func (h *PortfolioHandler) List(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	sortBy, err := SortBy(c)
	if err != nil {
		return nil, err
	}
	orderBy, err := OrderBy(c)
	if err != nil {
		return nil, err
	}

	list, err := service.PortfolioServiceInstance().GetPortfolioList(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": model.Process(list, sortBy, orderBy),
	}, nil
}

// Delete 删除自选股
func (h *PortfolioHandler) Delete(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	code, err := String(c, "code")
	if err != nil {
		return nil, serr.ErrBusiness("缺少参数:code")
	}
	sortBy, err := SortBy(c)
	if err != nil {
		return nil, err
	}
	orderBy, err := OrderBy(c)
	if err != nil {
		return nil, err
	}

	list, err := service.PortfolioServiceInstance().DeletePortfolio(ctx, uid, code)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": model.Process(list, sortBy, orderBy),
	}, nil
}

// Add 新增自选股
func (h *PortfolioHandler) Add(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	code, err := String(c, "code")
	if err != nil {
		return nil, serr.ErrBusiness("缺少参数:code")
	}
	sortBy, err := SortBy(c)
	if err != nil {
		return nil, err
	}
	orderBy, err := OrderBy(c)
	if err != nil {
		return nil, err
	}

	list, err := service.PortfolioServiceInstance().CreatePortfolio(ctx, uid, code)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": model.Process(list, sortBy, orderBy),
	}, nil
}
