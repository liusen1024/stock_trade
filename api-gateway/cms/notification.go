package handler

import (
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/util"
	"stock/common/timeconv"

	"github.com/gin-gonic/gin"
)

// LogHandler 管理
type LogHandler struct {
}

// NewLogHandler 单例
func NewLogHandler() *LogHandler {
	return &LogHandler{}
}

// Register 注册handler
func (h *LogHandler) Register(e *gin.Engine) {
	e.GET("/cms/log/sms", JSONWrapper(h.Sms))

}

type sms struct {
	ID       int64  `json:"id"`
	UserName string `json:"user_name"`
	Name     string `json:"name"`
	Agent    string `json:"agent"`
	Time     string `json:"time"`
	Content  string `json:"content"`
}

// Sms 日志-短信
func (h *LogHandler) Sms(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		BeginDate int32 `form:"begin_date" json:"begin_date"`
		EndDate   int32 `form:"end_date" json:"end_date"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	tx := db.StockDB().WithContext(ctx).Table("sms").Order("send_time desc")
	if req.BeginDate > 0 {
		tx.Where("send_time >= ?", timeconv.Int32ToTime(req.BeginDate).Format("2006-01-02"))
	}
	if req.EndDate > 0 {
		tx.Where("send_time <= ?", timeconv.Int32ToTime(req.EndDate).Format("2006-01-02"))
	}

	var total int64
	tx.Count(&total)

	var list []*model.Sms
	if err := tx.Find(&list).Error; err != nil {
		return nil, err
	}

	result := make([]*sms, 0)
	for _, it := range list {
		result = append(result, &sms{
			ID:       0,
			UserName: it.Phone,
			Name:     it.Phone,
			Agent:    "",
			Time:     it.Time.Format("2006-01-02 15:04:05"),
			Content:  it.Msg,
		})
	}
	return map[string]interface{}{
		"list":  result,
		"total": total,
	}, nil
}
