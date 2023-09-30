package handler

import (
	"stock/api-gateway/serr"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
	"stock/common/log"

	"github.com/gin-gonic/gin"
)

// SmsHandler 内容服务handler
type SmsHandler struct {
}

// NewSmsHandler 单例
func NewSmsHandler() *SmsHandler {
	return &SmsHandler{}
}

// Register 注册handler
func (h *SmsHandler) Register(e *gin.Engine) {
	// 发送短信
	e.GET("/sms/send", JSONWrapper(h.Send))
}

// Send 发送短信
func (h *SmsHandler) Send(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	phone, err := String(c, "phone")
	if err != nil {
		log.Errorf("Send err:%+v", err)
		return nil, err
	}
	// 简单手机号验证
	if len(phone) != 11 {
		log.Errorf("手机号码[%s]不是11位", phone)
		return nil, serr.New(serr.ErrCodeBusinessFail, "请输入正确手机号码")
	}

	err = service.SmsServiceInstance().Send(ctx, phone)
	if err != nil {
		log.Errorf("Send err:%+v", err)
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}
