package handler

import (
	"stock/api-gateway/dao"
	"stock/api-gateway/model"
	"stock/api-gateway/util"
	"stock/common/log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// SystemHandler 系统
type SystemHandler struct {
}

// NewSystemHandler 单例
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

type system struct {
	LimitPct               float64 `json:"limit_pct" form:"limit_pct"`                         // 涨跌幅买入限制
	CYBLimitPct            float64 `json:"cyb_limit_pct" form:"cyb_limit_pct"`                 // 创业板涨跌幅买入限制
	KCBLimitPct            float64 `json:"kcb_limit_pct" form:"cyb_limit_pct"`                 // 科创板涨跌幅买入限制
	STLimitPct             float64 `json:"st_limit_pct" form:"st_limit_pct"`                   // ST涨跌幅买入限制
	IsSupportSTStock       bool    `json:"st_forbid" form:"st_forbid"`                         // ST股是否允许交易:true允许交易,false不允许交易
	ClosePct               float64 `json:"close_pct" form:"close_pct"`                         // 平仓线率
	WarnPct                float64 `json:"warn_pct" form:"warn_pct"`                           // 警戒线率
	IsSupportKCBBoard      bool    `json:"sge_board_forbid" form:"sge_board_forbid"`           // 科创板是否允许交易:true允许交易,false不允许交易
	BuyFee                 float64 `json:"buy_fee" form:"buy_fee"`                             // 买入手续费
	MiniChargeFee          float64 `json:"mini_charge_fee" form:"mini_charge_fee"`             // 最低手续费
	RegistCode             bool    `json:"regist_code" form:"regist_code"`                     // 注册须推荐码:true必须填写正确推荐码
	WithdrawBeginTime      string  `json:"withdraw_begin_time" form:"withdraw_begin_time"`     // 提现开始时间
	WithdrawEndTime        string  `json:"withdraw_end_time" form:"withdraw_end_time"`         // 提现结束时间
	RechargeNotice         bool    `json:"recharge_notice" form:"recharge_notice"`             // 用户充值短信通知管理:true通知,false不通知
	RegisterNotice         bool    `json:"register_notice" form:"register_notice"`             // 用户注册短信通知管理:true通知,false不通知
	WithdrawNotice         bool    `json:"withdraw_notice" form:"withdraw_notice"`             // 用户提现通知管理:true通知,false不通知
	Broker                 bool    `json:"broker" form:"broker"`                               // 是否对接券商:true对接,false不对接
	WarnCanBuy             bool    `json:"warn_can_buy" form:"warn_can_buy"`                   // 触发警戒线允许买入:true允许买入,false不允许买入
	IsSupportCYBBoard      bool    `json:"cyb_board_forbid" form:"cyb_board_forbid"`           // 创业板允许交易:true允许交易,false不允许交易
	SellFee                float64 `json:"sell_fee" form:"sell_fee"`                           // 卖出手续费
	SingleBuyPct           float64 `json:"single_buy_pct" form:"single_buy_pct"`               // 单只股票最大持仓比率
	HolidayCharge          bool    `json:"holiday_charge" form:"holiday_charge"`               // 节假日收取管理费:true节假日收取留仓费,false不收取
	BankName               string  `json:"bank_name" form:"bank_name"`                         // 收款人姓名
	BankNo                 string  `json:"bank_no" form:"bank_no"`                             // 收款银行卡号
	BankAddr               string  `json:"bank_addr" form:"bank_addr"`                         // 收款行地址
	BankChannel            bool    `json:"bank_channel" form:"bank_channel"`                   // 银行卡收款渠道
	QRCodeChannel          bool    `json:"qrcode_channel" form:"qrcode_channel"`               // 二维码收款渠道
	AlipayChannel          bool    `json:"alipay_channel" form:"alipay_channel"`               // 支付宝H5渠道
	IsSupportDayContract   bool    `json:"contract_day_status" form:"contract_day_status"`     // 按天合约类型
	IsSupportWeekContract  bool    `json:"contract_week_status" form:"contract_week_status"`   // 按周合约类型
	IsSupportMonthContract bool    `json:"contract_month_status" form:"contract_month_status"` // 按月合约类型
	ContractDayFee         float64 `json:"contract_day_fee" form:"contract_day_fee"`           // 按天管理费率
	ContractWeekFee        float64 `json:"contract_week_fee" form:"contract_week_fee"`         // 按周管理费率
	ContractMonthFee       float64 `json:"contract_month_fee" form:"contract_month_fee"`       // 按月管理费率
	ContractLever          []int64 `json:"contract_lever" form:"contract_lever"`               // 合约杠杆
	AdminPhone             string  `json:"admin_phone" form:"admin_phone"`                     // 管理员手机号
}

// Register 注册handler
func (h *SystemHandler) Register(e *gin.Engine) {
	// 股票列表
	e.GET("/cms/system/get", JSONWrapper(h.Get))
	e.POST("/cms/system/set", JSONWrapper(h.Set))
}

// Set 设置
func (h *SystemHandler) Set(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	var req system
	if err := c.Bind(&req); err != nil {
		return nil, err
	}
	levers := make([]string, 0)
	for _, it := range req.ContractLever {
		levers = append(levers, strconv.FormatInt(it, 10))
	}
	if err := dao.SysDaoInstance().Update(ctx, &model.SysParam{
		StartWithdrawTime:      req.WithdrawBeginTime,
		StopWithdrawTime:       req.WithdrawEndTime,
		LimitPct:               req.LimitPct,
		CYBLimitPct:            req.CYBLimitPct,
		STLimitPct:             req.STLimitPct,
		IsSupportSTStock:       req.IsSupportSTStock,
		IsSupportKCBBoard:      req.IsSupportKCBBoard,
		IsSupportCYBBoard:      req.IsSupportCYBBoard,
		IsSupportBJBoard:       false,
		SinglePositionPct:      req.SingleBuyPct,
		RechargeNotice:         req.RechargeNotice,
		RegisterNotice:         req.RegisterNotice,
		WithdrawNotice:         req.WithdrawNotice,
		WarnPct:                req.WarnPct,
		LowWarnCanBuy:          req.WarnCanBuy,
		ClosePct:               req.ClosePct,
		HolidayCharge:          req.HolidayCharge,
		BuyFee:                 req.BuyFee,
		SellFee:                req.SellFee,
		RegistCode:             req.RegistCode,
		BankNo:                 req.BankNo,
		BankName:               req.BankName,
		BankAddr:               req.BankAddr,
		BankChannel:            req.BankChannel,
		QrcodeChannel:          req.QRCodeChannel,
		AlipayChannel:          req.AlipayChannel,
		ContractLever:          strings.Join(levers, ","),
		IsSupportDayContract:   req.IsSupportDayContract,
		IsSupportWeekContract:  req.IsSupportWeekContract,
		IsSupportMonthContract: req.IsSupportMonthContract,
		DayContractFee:         req.ContractDayFee,
		WeekContractFee:        req.ContractWeekFee,
		MonthContractFee:       req.ContractMonthFee,
		MiniChargeFee:          req.MiniChargeFee,
		IsSupportBroker:        req.Broker,
		AdminPhone:             req.AdminPhone,
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"result": true,
	}, nil
}

// Get 查询
func (h *SystemHandler) Get(c *gin.Context) (interface{}, error) {
	ctx := util.RPCContext(c)
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		return &system{}, nil
	}
	contractLevers := make([]int64, 0)
	segs := strings.Split(sys.ContractLever, ",")
	for _, seg := range segs {
		lever, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			log.Errorf("parse err :%+v", err)
			continue
		}
		contractLevers = append(contractLevers, lever)
	}
	return &system{
		LimitPct:               sys.LimitPct,
		CYBLimitPct:            sys.CYBLimitPct,
		STLimitPct:             sys.STLimitPct,
		IsSupportSTStock:       sys.IsSupportSTStock,
		ClosePct:               sys.ClosePct,
		WarnPct:                sys.WarnPct,
		IsSupportKCBBoard:      sys.IsSupportKCBBoard,
		BuyFee:                 sys.BuyFee,
		MiniChargeFee:          sys.MiniChargeFee,
		RegistCode:             sys.RegistCode,
		WithdrawBeginTime:      sys.StartWithdrawTime,
		WithdrawEndTime:        sys.StopWithdrawTime,
		RechargeNotice:         sys.RechargeNotice,
		RegisterNotice:         sys.RegisterNotice,
		WithdrawNotice:         sys.WithdrawNotice,
		Broker:                 sys.IsSupportBroker,
		WarnCanBuy:             sys.LowWarnCanBuy,
		IsSupportCYBBoard:      sys.IsSupportCYBBoard,
		SellFee:                sys.SellFee,
		SingleBuyPct:           sys.SinglePositionPct,
		HolidayCharge:          sys.HolidayCharge,
		BankName:               sys.BankName,
		BankNo:                 sys.BankNo,
		BankAddr:               sys.BankAddr,
		BankChannel:            sys.BankChannel,
		QRCodeChannel:          sys.QrcodeChannel,
		AlipayChannel:          sys.AlipayChannel,
		IsSupportDayContract:   sys.IsSupportDayContract,
		IsSupportWeekContract:  sys.IsSupportWeekContract,
		IsSupportMonthContract: sys.IsSupportMonthContract,
		ContractDayFee:         sys.DayContractFee,
		ContractWeekFee:        sys.WeekContractFee,
		ContractMonthFee:       sys.MonthContractFee,
		ContractLever:          contractLevers,
		AdminPhone:             sys.AdminPhone,
	}, nil
}
