package handler

import (
	"bufio"
	"io"
	"os"
	"stock/api-gateway/dao"
	"stock/api-gateway/serr"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
	"stock/common/log"

	"github.com/gin-gonic/gin"
)

// MyHandler 首页服务handler
type MyHandler struct {
}

// NewMyHandler 单例
func NewMyHandler() *MyHandler {
	return &MyHandler{}
}

// Register 注册handler
func (h *MyHandler) Register(e *gin.Engine) {
	// 我的
	e.GET("/my", JSONWrapper(h.My))
	// 我的消息
	e.GET("/my/msg", JSONWrapper(h.Msg))
	// 实名认证,初始化
	e.GET("/my/authentication/get", JSONWrapper(h.GetAuthentication))
	// 实名认证
	e.GET("/my/authentication", JSONWrapper(h.Authentication))
	// 资金明细
	e.GET("/my/balance", JSONWrapper(h.Balance))
	// 充值页面初始化
	e.GET("/my/recharge/get", JSONWrapper(h.GetRecharge))
	// 充值-支付宝
	e.GET("/my/recharge/alipay", PureJSONWrapper(h.RechargeAlipay))
	// 支付宝回调
	e.POST("/alipay/callback", JSONWrapper(h.AliPayNotify))
	// 银行卡充值,获取充值银行卡的账号信息
	e.GET("/my/recharge/bank", JSONWrapper(h.RechargeBank))
	// 银行卡充值,提交
	e.GET("/my/recharge/bank/commit", JSONWrapper(h.RechargeBankCommit))
	// 扫码支付初始化
	e.GET("/my/recharge/qrcode", JSONWrapper(h.RechargeQrcode))
	// 扫码支付提交
	e.GET("/my/recharge/qrcode/commit", JSONWrapper(h.RechargeQrcodeCommit))
	// 提现初始化
	e.GET("/my/withdraw", JSONWrapper(h.Withdraw))
	// 提现提交
	e.GET("/my/withdraw/commit", JSONWrapper(h.WithdrawCommit))
	// 交易规则
	e.GET("/my/rule", PureJSONWrapper(h.Rule))
	// 联系客服
	e.GET("/my/customer", JSONWrapper(h.Customer))
}

// Customer 联系客服
func (h *MyHandler) Customer(c *gin.Context) (interface{}, error) {
	return map[string]interface{}{
		"img": "",
	}, nil
}

// Rule 交易规则
func (h *MyHandler) Rule(c *gin.Context) (interface{}, error) {
	file, err := os.Open("./rule.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	list := make([]string, 0)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		list = append(list, string(line))
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

func (h *MyHandler) WithdrawCommit(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	money, err := Float64(c, "money")
	if money < 0.01 {
		log.Errorf("转出金额错误:%+v", err)
		return nil, serr.ErrBusiness("转出金额错误")
	}
	name, err := String(c, "name")
	if err != nil {
		return nil, serr.ErrBusiness("请填写正确的收款人")
	}
	bankNo, err := String(c, "bank_no")
	if err != nil {
		return nil, serr.ErrBusiness("请填写正确的银行卡号")
	}
	code, err := String(c, "code")
	if err != nil {
		return nil, serr.ErrBusiness("请填写正确的验证码")
	}
	if err := service.MyServiceInstance().WithdrawCommit(ctx, uid, money, name, bankNo, code); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// Withdraw 提现初始化
func (h *MyHandler) Withdraw(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	name, bankNo, money, err := service.MyServiceInstance().Withdraw(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"name":    name,   // 收款人姓名
		"bank_no": bankNo, // 收款人银行卡
		"money":   money,  // 可提现金额
	}, nil
}

// RechargeQrcodeCommit 扫码支付提交
func (h *MyHandler) RechargeQrcodeCommit(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	money, _ := Float64(c, "money")
	if money < 0.001 {
		return nil, serr.ErrBusiness("转入金额错误")
	}
	orderNo, err := String(c, "order_no")
	if err != nil {
		return nil, err
	}
	if err := service.MyServiceInstance().RechargeQrcodeCommit(ctx, uid, money, orderNo); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// RechargeQrcode 扫码支付初始化
func (h *MyHandler) RechargeQrcode(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	money, _ := Float64(c, "money")
	if money < 0.001 {
		return nil, serr.ErrBusiness("转入金额错误")
	}
	img, orderNo, err := service.MyServiceInstance().RechargeQrcode(ctx, uid, money)
	return map[string]interface{}{
		"img":      img,
		"order_no": orderNo,
	}, nil
}

func (h *MyHandler) RechargeBankCommit(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	money, _ := Float64(c, "money")
	if money < 0.001 {
		return nil, serr.ErrBusiness("转入金额错误")
	}
	if err := service.MyServiceInstance().RechargeBankCommit(ctx, uid, money); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

func (h *MyHandler) RechargeBank(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	_, err := UserID(c)
	if err != nil {
		return nil, err
	}
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	return map[string]interface{}{
		"bank_no": sys.BankNo,
		"name":    sys.BankName,
		"address": sys.BankAddr,
	}, nil
}

func (h *MyHandler) AliPayNotify(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	type request struct {
		AppID       string  `form:"app_id" json:"app_id"`
		TradeNo     string  `form:"out_trade_no" json:"out_trade_no"`
		Buyer       string  `form:"buyer_logon_id" json:"buyer_logon_id"`
		TradeStatus string  `form:"trade_status" json:"trade_status"`
		TotalAmount float64 `form:"total_amount" json:"total_amount"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		log.Errorf("绑定参数失败:%+v", err)
		return nil, err
	}
	log.Infof("%+v", req)
	if req.TradeStatus != "TRADE_SUCCESS" {
		return nil, nil
	}
	service.AlipayServiceInstance().AliPayNotify(ctx, req.TradeNo, req.Buyer, req.TotalAmount)
	return "success", nil
}

func (h *MyHandler) RechargeAlipay(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	money, _ := Float64(c, "money")
	if money < 0.01 {
		return nil, serr.ErrBusiness("转入金额错误")
	}
	url, err := service.MyServiceInstance().RechargeAlipay(ctx, uid, money)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"url": url,
	}, nil
}

func (h *MyHandler) GetRecharge(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"bank":   sys.BankChannel,
		"alipay": sys.AlipayChannel,
		"qrcode": sys.QrcodeChannel,
	}, nil
}

// Balance 资金明细
func (h *MyHandler) Balance(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	list, balance, err := service.MyServiceInstance().Balance(ctx, uid)
	return map[string]interface{}{
		"list":    list,
		"balance": balance,
	}, nil
}

func (h *MyHandler) Authentication(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	name, err := String(c, "name")
	if err != nil {
		return nil, err
	}
	idNo, err := String(c, "id_no")
	if err != nil {
		return nil, err
	}
	if err := service.MyServiceInstance().Authentication(ctx, uid, name, idNo); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

func (h *MyHandler) GetAuthentication(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	user, err := dao.UserDaoInstance().GetUserByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"name":  user.Name,
		"id_no": user.ICCID,
	}, nil
}

func (h *MyHandler) Msg(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	list, err := service.MyServiceInstance().Msg(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// My 我的页面首页
func (h *MyHandler) My(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	return service.MyServiceInstance().My(ctx, uid)
}
