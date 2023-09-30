package handler

import (
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/service"
	"stock/api-gateway/util"

	"github.com/gin-gonic/gin"
)

// ContractHandler 首页服务handler
type ContractHandler struct {
}

// NewContractHandler 单例
func NewContractHandler() *ContractHandler {
	return &ContractHandler{}
}

// Register 注册handler
func (h *ContractHandler) Register(e *gin.Engine) {
	// 合约查询
	e.GET("/contract/list", JSONWrapper(h.List))
	// 申请合约初始化
	e.GET("/contract/apply_init", JSONWrapper(h.ApplyInit))
	// 申请合约-立即体验
	e.GET("/contract/apply", JSONWrapper(h.ContractApply))
	// 申请合约-确认订单
	e.GET("/contract/create", JSONWrapper(h.Create))
	// 申请合约-详情
	e.GET("/contract/detail", JSONWrapper(h.Detail))
	// 合约结算
	e.GET("/contract/close", JSONWrapper(h.Close))
	// 追加保证金页面初始化
	e.GET("/contract/get_append_money", JSONWrapper(h.GetAppendMoney))
	// 追加保证金
	e.GET("/contract/append_money", JSONWrapper(h.AppendMoney))
	// 扩大合约初始化
	e.GET("/contract/get_expand_money", JSONWrapper(h.GetExpandMoney))
	// 扩大合约
	e.GET("/contract/expand_money", JSONWrapper(h.ExpandMoney))
	// 历史合约
	e.GET("/contract/history", JSONWrapper(h.HisContract))
	// 合约列表
	e.GET("/contract/get", JSONWrapper(h.GetContract))
	// 合约选中
	e.GET("/contract/select", JSONWrapper(h.Select))
	// 查询合约提盈
	e.GET("/contract/get_withdraw_profit", JSONWrapper(h.GetWithdrawProfit))
	// 合约提盈
	e.GET("/contract/withdraw_profit", JSONWrapper(h.WithdrawProfit))
}

func (h *ContractHandler) List(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	result, err := service.ContractServiceInstance().List(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": result,
	}, nil
}

// ApplyInit 申请合约初始化配置
func (h *ContractHandler) ApplyInit(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	return service.ContractServiceInstance().ApplyInit(ctx, uid)
}

// ContractApply 申请合约-立即体验(创建合约)
func (h *ContractHandler) ContractApply(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	type request struct {
		Money         float64 `form:"money"`
		ContractType  int64   `form:"type" json:"type"`
		ContractLever int64   `form:"lever" json:"lever"`
	}
	var req request
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	if req.Money < 0 {
		return nil, serr.ErrBusiness("申请资金不合法,请输入正确资金")
	}
	if _, ok := model.ContractTypeMap[req.ContractType]; !ok {
		return nil, serr.ErrBusiness("申请合约类型不存在")
	}
	return service.ContractServiceInstance().ContractApply(ctx, uid, req.Money, req.ContractType, req.ContractLever)
}

// Create 确认合约
func (h *ContractHandler) Create(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	err = service.ContractServiceInstance().Create(ctx, uid, contractID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// Detail 合约明细
func (h *ContractHandler) Detail(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	return service.ContractServiceInstance().Detail(ctx, contractID)
}

// Close 合约结算
func (h *ContractHandler) Close(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	if err := service.ContractServiceInstance().Settlement(ctx, contractID); err != nil {
		return nil, err
	}
	return map[string]bool{
		"result": true,
	}, nil
}

// GetAppendMoney 追加保证金页面初始化
func (h *ContractHandler) GetAppendMoney(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	return service.ContractServiceInstance().GetAppendMoney(ctx, contractID)
}

// AppendMoney 追加保证金
func (h *ContractHandler) AppendMoney(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	money, err := Float64(c, "money")
	if err != nil {
		return nil, serr.ErrBusiness("请输入追加金额")
	}
	if err := service.ContractServiceInstance().AppendMoney(ctx, contractID, money); err != nil {
		return nil, err
	}
	return map[string]bool{
		"result": true,
	}, nil
}

// GetExpandMoney 扩大合约页面初始化
func (h *ContractHandler) GetExpandMoney(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	return service.ContractServiceInstance().GetAppendMoney(ctx, contractID)
}

// ExpandMoney 扩大资金
func (h *ContractHandler) ExpandMoney(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	money, err := Float64(c, "money")
	if err != nil {
		return nil, serr.ErrBusiness("请输入金额")
	}
	if err := service.ContractServiceInstance().ExpandMoney(ctx, contractID, money); err != nil {
		return nil, err
	}
	return map[string]bool{
		"result": true,
	}, nil
}

// HisContract 历史合约
func (h *ContractHandler) HisContract(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	list, err := service.ContractServiceInstance().HisContract(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// GetContract 切换合约时返回的合约列表
func (h *ContractHandler) GetContract(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	list, err := service.ContractServiceInstance().GetContract(ctx, uid)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// Select 选中合约
func (h *ContractHandler) Select(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	if err := service.ContractServiceInstance().Select(ctx, uid, contractID); err != nil {
		return nil, err
	}
	return map[string]bool{
		"result": true,
	}, nil
}

// GetWithdrawProfit 合约提盈初始化
// 提盈策略：1. 空仓 && 2.保证金大于原始资金
func (h *ContractHandler) GetWithdrawProfit(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}

	return service.ContractServiceInstance().GetWithdrawProfit(ctx, contractID)
}

// WithdrawProfit 合约提盈
func (h *ContractHandler) WithdrawProfit(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	Money, err := Float64(c, "money")
	if err != nil {
		return nil, serr.ErrBusiness("请输入正确的金额")
	}
	if Money < 0 {
		return nil, serr.ErrBusiness("提取失败:金额错误")
	}
	if err := service.ContractServiceInstance().WithdrawProfit(ctx, contractID, Money); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}
