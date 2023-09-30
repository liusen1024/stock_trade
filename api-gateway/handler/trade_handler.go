package handler

import (
	"github.com/gin-gonic/gin"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
)

// TradeHandler 内容服务handler
type TradeHandler struct {
}

// NewTradeHandler 单例
func NewTradeHandler() *TradeHandler {
	return &TradeHandler{}
}

// Register 注册handler
func (s *TradeHandler) Register(e *gin.Engine) {
	// 交易页面初始化
	e.GET("/trade/init", JSONWrapper(s.InitTrade))
	// 查询持仓明细
	e.GET("/trade/position_detail", JSONWrapper(s.PositionDetail))
	// 查询今日成交
	e.GET("/trade/today", JSONWrapper(s.TodayDeal))
	// 查询历史成交
	e.GET("/trade/history", JSONWrapper(s.HistoryDeal))
	// 查询费用单
	e.GET("/trade/contract_fee", JSONWrapper(s.ContractFee))
	// 成交明细
	e.GET("/trade/detail", JSONWrapper(s.TradeDetail))
	// 查询持仓
	e.GET("/trade/position", JSONWrapper(s.StockPosition))
	// 买入界面初始化
	e.GET("/trade/init_buy", JSONWrapper(s.InitBuy))
	// 买入
	e.GET("/trade/buy", JSONWrapper(s.Buy))
	// 卖出界面初始化
	e.GET("/trade/init_sell", JSONWrapper(s.InitSell))
	// 卖出
	e.GET("/trade/sell", JSONWrapper(s.Sell))
	// 撤单界面初始化查询
	e.GET("/trade/init_withdraw", JSONWrapper(s.InitWithdraw))
	// 撤单
	e.GET("/trade/withdraw", JSONWrapper(s.Withdraw))
	// 查询-委托记录
	e.GET("/trade/entrust/list", JSONWrapper(s.GetEntrustList))
	// 查询-已清仓股票
	e.GET("/trade/sell_out", JSONWrapper(s.GetSellOut))
}

// GetSellOut 查询-已清仓股票
func (h *TradeHandler) GetSellOut(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	list, err := service.TradeServiceInstance().GetSellOut(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// GetEntrustList 查询-委托记录
func (h *TradeHandler) GetEntrustList(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	list, err := service.TradeServiceInstance().GetEntrustList(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// Withdraw 撤单操作
func (h *TradeHandler) Withdraw(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	entrustID, err := EntrustID(c)
	if err != nil {
		return nil, err
	}
	if err := service.TradeServiceInstance().Withdraw(ctx, entrustID); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// InitWithdraw 撤单
func (h *TradeHandler) InitWithdraw(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	_, err := UserID(c)
	if err != nil {
		return nil, err
	}
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	list, err := service.TradeServiceInstance().InitWithdraw(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// InitTrade 初始化交易页面
func (s *TradeHandler) InitTrade(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	contractID, _ := ContractID(c)
	return service.TradeServiceInstance().InitTrade(ctx, uid, contractID)
}

// PositionDetail 持仓明细
func (s *TradeHandler) PositionDetail(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	positionID, err := Int64(c, "position_id")
	if err != nil {
		return nil, err
	}
	return service.TradeServiceInstance().PositionDetail(ctx, positionID)
}

// InitBuy 买入界面初始化
func (s *TradeHandler) InitBuy(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	code, err := StockCode(c)
	if err != nil {
		return nil, err
	}
	return service.TradeServiceInstance().InitBuy(ctx, uid, contractID, code)
}

// InitSell 卖出界面初始化
func (s *TradeHandler) InitSell(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	_, err := UserID(c)
	if err != nil {
		return nil, err
	}
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	code, err := StockCode(c)
	if err != nil {
		return nil, err
	}
	return service.TradeServiceInstance().InitSell(ctx, contractID, code)
}

// Buy 买入
func (s *TradeHandler) Buy(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c) // 用户ID
	if err != nil {
		return nil, err
	}
	contractID, err := ContractID(c) // 合约ID
	if err != nil {
		return nil, err
	}
	code, err := StockCode(c) // 股票代码
	if err != nil {
		return nil, err
	}
	price, err := Price(c) // 股票价格
	if err != nil {
		price = 0
	}
	amount, err := Amount(c) // 股票数量
	if err != nil {
		return nil, err
	}
	if amount%100 != 0 {
		return nil, serr.New(serr.ErrCodeBusinessFail, "委托失败:交易股数应为100整倍数")
	}
	entrustProp, err := EntrustProp(c) // 委托类型
	if err != nil {
		return nil, err
	}
	if entrustProp == model.EntrustPropTypeLimitPrice && price == 0 {
		return nil, serr.New(serr.ErrCodeBusinessFail, "委托失败:委托价格错误")
	}
	if err := service.TradeServiceInstance().Buy(ctx, &model.EntrustPackage{
		UID:         uid,         // 用户UID
		ContractID:  contractID,  // 合约ID
		Code:        code,        // 股票代码
		Price:       price,       // 股票价格
		Amount:      amount,      // 股票数量
		EntrustProp: entrustProp, // 委托类型:1限价 2市价
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// Sell 卖出
func (s *TradeHandler) Sell(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c) // 用户ID
	if err != nil {
		return nil, err
	}
	contractID, err := ContractID(c) // 合约ID
	if err != nil {
		return nil, err
	}
	code, err := StockCode(c) // 股票代码
	if err != nil {
		return nil, err
	}
	price, err := Price(c) // 股票价格
	if err != nil {
		return nil, err
	}
	amount, err := Amount(c) // 股票数量
	if err != nil {
		return nil, err
	}
	if amount%100 != 0 && !service.TradeServiceInstance().HoldZeroShare(ctx, contractID) {
		return nil, serr.New(serr.ErrCodeBusinessFail, "委托失败:交易股数应为100整倍数")
	}

	entrustProp, err := EntrustProp(c) // 委托类型
	if err != nil {
		return nil, err
	}
	if entrustProp == model.EntrustPropTypeLimitPrice && price == 0 {
		return nil, serr.New(serr.ErrCodeBusinessFail, "委托失败:委托价格错误")
	}

	if err := service.TradeServiceInstance().Sell(ctx, &model.EntrustPackage{
		UID:         uid,            // 用户UID
		ContractID:  contractID,     // 合约ID
		Code:        code,           // 股票代码
		Price:       price,          // 股票价格
		Amount:      amount,         // 股票数量
		EntrustProp: entrustProp,    // 委托类型:1限价 2市价
		Mode:        model.UserMode, // 委托方世:userMode用户主动委托,systemMode系统委托
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// StockPosition 查询持仓
func (s *TradeHandler) StockPosition(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	uid, err := UserID(c)
	if err != nil {
		return nil, err
	}
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	return service.TradeServiceInstance().Position(ctx, uid, contractID)
}

// TodayDeal 今日成交
func (s *TradeHandler) TodayDeal(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	_, err := UserID(c)
	if err != nil {
		return nil, err
	}
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	list, err := service.TradeServiceInstance().TodayDeal(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// HistoryDeal 今日成交
func (s *TradeHandler) HistoryDeal(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	list, err := service.TradeServiceInstance().HistoryDeal(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// ContractFee 合约费用单
func (s *TradeHandler) ContractFee(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	contractID, err := ContractID(c)
	if err != nil {
		return nil, err
	}
	list, err := service.TradeServiceInstance().ContractFee(ctx, contractID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"list": list,
	}, nil
}

// TradeDetail 交易明细
func (s *TradeHandler) TradeDetail(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	entrustID, err := Int64(c, "id")
	if err != nil {
		return nil, err
	}
	return service.TradeServiceInstance().TradeDetail(ctx, entrustID)
}
