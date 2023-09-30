package handler

import (
	"stock/api-gateway/model"
	"stock/api-gateway/service"
	"stock/api-gateway/util"

	"github.com/gin-gonic/gin"
)

// MarketHandler 市场handler
type MarketHandler struct {
}

// NewMarketHandler 单例
func NewMarketHandler() *MarketHandler {
	return &MarketHandler{}
}

// Register 注册handler
func (h *MarketHandler) Register(e *gin.Engine) {
	// 市场首页
	e.GET("/market/get", PureJSONWrapper(h.Market))
	// 市场-获取全部行业
	e.GET("/market/all_industry", JSONWrapper(h.GetAllIndustry))
	// 市场-获取全部概念
	e.GET("/market/all_concept", JSONWrapper(h.GetAllConcept))
	// 市场-板块股票列表
	e.GET("/market/bk_detail", JSONWrapper(h.GetBKDetail))
}

// GetAllConcept 获取所有概念版块
func (h *MarketHandler) GetAllConcept(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	list, err := service.MarketServiceInstance().GetAllConcept(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// GetBKDetail 获取单独的行业
func (h *MarketHandler) GetBKDetail(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	sectorCode, err := String(c, "code")
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
	list, err := service.MarketServiceInstance().GetBKDetail(ctx, sectorCode)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": model.Process(list, sortBy, orderBy),
	}, nil
}

// GetAllIndustry 获取所有行业
func (h *MarketHandler) GetAllIndustry(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	list, err := service.MarketServiceInstance().GetAllIndustry(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// Market 市场列表
func (h *MarketHandler) Market(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	result, err := service.MarketServiceInstance().GetMarket(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}
