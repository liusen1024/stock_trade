package handler

import (
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/service"
	"stock/api-gateway/util"

	"github.com/gin-gonic/gin"
)

// StockHandler 股票管理
type StockHandler struct {
}

// NewStockHandler 单例
func NewStockHandler() *StockHandler {
	return &StockHandler{}
}

// Register 注册handler
func (h *StockHandler) Register(e *gin.Engine) {
	// 股票列表
	e.GET("/cms/stock/list", JSONWrapper(h.StockList))
	e.POST("/cms/stock/update", JSONWrapper(h.UpdateStatus))
}

// UpdateStatus 更新股票状态
func (h *StockHandler) UpdateStatus(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		ID     int64 `form:"user_name" json:"id"`
		Status bool  `form:"status" json:"status"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	var status int64
	if req.Status {
		status = model.StockDataStatusDisable
	} else {
		status = model.StockDataStatusEnable
	}
	if err := service.StockDataServiceInstance().UpdateStatusByID(ctx, req.ID, status); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// StockList 股票列表
func (h *StockHandler) StockList(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		StockCode string `form:"stock_code" json:"stock_code"`
		Type      int64  `form:"type" json:"type"` // 0全部 1 是 2不是
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	tx := db.StockDB().WithContext(ctx).Table("stock_data")
	if len(req.StockCode) > 0 {
		tx.Where("code = ?", req.StockCode)
	}
	if req.Type == 1 {
		tx.Where("status = ?", model.StockDataStatusEnable)
	} else if req.Type == 2 {
		tx.Where("status = ?", model.StockDataStatusDisable)
	}

	var stocks []*model.StockData
	if err := tx.Find(&stocks).Error; err != nil {
		return nil, err
	}
	list := make([]*model.CmsStockDataResp, 0)
	for _, it := range stocks {
		list = append(list, &model.CmsStockDataResp{
			ID:        it.ID,
			StockCode: it.Code,
			StockName: it.Name,
			IPODate:   it.IPODay.Format("2006-01-02"),
			Status:    it.Status == 1,
		})
	}

	// 下载则下发文件
	if IsDownload(c) {
		var res []interface{}
		for _, it := range list {
			res = append(res, it)
		}
		Download(c, []string{
			"ID", "股票代码", "股票名称", "上市日期", "是否可买:true可买,false不可买",
		}, res)
	}

	count := len(list)
	start, end := SlicePage(c, count)

	return map[string]interface{}{
		"list":  list[start:end],
		"total": count,
	}, nil
}
