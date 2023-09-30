package handler

import (
	"stock/api-gateway/serr"
	"stock/api-gateway/service"
	"stock/api-gateway/util"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户服务
type UserHandler struct {
}

// NewUserHandler 单例
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Register 注册handler
func (h *UserHandler) Register(e *gin.Engine) {
	// 登录
	e.GET("/user/login", JSONWrapper(h.Login))
	// 注册
	e.GET("/user/register", JSONWrapper(h.RegisterUser))
	// 找回密码
	e.GET("/user/update_password", JSONWrapper(h.UpdatePassword))
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type loginReq struct {
		UserName string `form:"user_name"`
		Password string `form:"password"`
	}
	var req loginReq
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.UserName) == 0 {
		return nil, serr.ErrBusiness("请输入账号")
	}
	if len(req.Password) == 0 {
		return nil, serr.ErrBusiness("请输入密码")
	}

	user, err := service.UserServiceInstance().Login(ctx, req.UserName, req.Password)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"uid": user.ID,
	}, nil
}

// RegisterUser 用户注册
func (h *UserHandler) RegisterUser(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type registerReq struct {
		UserName     string `form:"user_name"`
		Password     string `form:"password"`
		Code         string `form:"sms_code"`
		RegisterCode string `form:"register_code" json:"register_code"`
	}
	var req registerReq
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.UserName) == 0 {
		return nil, serr.ErrBusiness("请输入注册手机号")
	}
	if len(req.Password) == 0 {
		return nil, serr.ErrBusiness("请输入密码")
	}
	if len(req.Code) == 0 {
		return nil, serr.ErrBusiness("请输入验证码")
	}
	if len(req.Password) < 6 {
		return nil, serr.ErrBusiness("请重新设置密码，密码长度不低于6位")
	}
	if err := service.UserServiceInstance().RegisterUser(ctx, req.UserName, req.Password, req.Code, req.RegisterCode); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil

}

// UpdatePassword 重新设定密码(找回密码)
func (h *UserHandler) UpdatePassword(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type resetReq struct {
		UserName string `form:"user_name"`
		Password string `form:"password"`
		Code     string `form:"sms_code"`
	}
	var req resetReq
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if len(req.UserName) == 0 {
		return nil, serr.ErrBusiness("请输入手机号码")
	}
	if len(req.Password) == 0 {
		return nil, serr.ErrBusiness("请输入新密码")
	}
	if len(req.Code) == 0 {
		return nil, serr.ErrBusiness("请输入验证码")
	}
	if len(req.Password) < 6 {
		return nil, serr.ErrBusiness("新密码长度大于6位")
	}
	if err := service.UserServiceInstance().UpdatePassword(ctx, req.UserName, req.Password, req.Code); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil

}
