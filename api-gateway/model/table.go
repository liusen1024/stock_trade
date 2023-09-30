package model

import (
	"time"
)

// /////////////////////////////////users表///////////////////////////////////
const (
	UserStatusActive = 1 // 激活状态
	UserStatusFrezze = 2 // 冻结状态
)

// User 用户信息
type User struct {
	ID                int64     `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	UserName          string    `gorm:"column:user_name"`
	Password          string    `gorm:"column:password"`
	Status            int8      `gorm:"column:status"`                           // 1:激活 2:冻结
	CurrentContractID int64     `gorm:"column:current_contract_id"`              // 当前合约ID
	Name              string    `gorm:"column:name"`                             // 姓名
	ICCID             string    `gorm:"column:icc_id"`                           // 身份证号码
	BankNumber        string    `gorm:"column:bank_number"`                      // 银行卡号码
	RoleID            int64     `gorm:"column:role_id"`                          // 代理商
	CreateAt          time.Time `gorm:"column:created_at"`                       // 创建时间
	Money             float64   `gorm:"column:money"`                            // 保证金
	FreezeMoney       float64   `gorm:"column:freeze_money" json:"freeze_money"` // 冻结资金
}

///////////////////////////////////users表///////////////////////////////////

///////////////////////////////////sysParam表///////////////////////////////////

// SysParam 系统参数表
type SysParam struct {
	StartWithdrawTime      string  `gorm:"column:start_withdraw_time"`         // 提现开始时间
	StopWithdrawTime       string  `gorm:"column:stop_withdraw_time"`          // 提现结束时间
	LimitPct               float64 `gorm:"column:limit_pct"`                   // 涨跌幅限制:股票涨跌达到涨幅限制买入
	CYBLimitPct            float64 `gorm:"column:cyb_limit_pct"`               // 创业板涨跌幅限制
	KCBLimitPct            float64 `json:"kcb_limit_pct" form:"cyb_limit_pct"` // 科创板涨跌幅买入限制
	STLimitPct             float64 `gorm:"column:st_limit_pct"`                // ST涨跌幅限制:股票涨跌达到涨幅限制买入
	IsSupportSTStock       bool    `gorm:"column:is_support_st_stock"`         // true:禁止st股票交易 false:允许st股交易
	IsSupportKCBBoard      bool    `gorm:"column:is_support_sge_board"`        // 科创板允许交易:true禁止 false允许
	IsSupportCYBBoard      bool    `gorm:"column:is_support_cyb_board"`        // 创业板允许交易:true禁止 false允许
	IsSupportBJBoard       bool    `gorm:"column:is_support_bj_board"`         // 北交所允许交易:true禁止 false允许
	SinglePositionPct      float64 `gorm:"column:single_position_pct"`         // 单只股票最大持仓比率
	RechargeNotice         bool    `gorm:"column:recharge_notice"`             // 充值通知管理员:true通知,false不通知
	RegisterNotice         bool    `gorm:"column:register_notice"`             // 注册通知管理员:true通知,false不通知
	WithdrawNotice         bool    `gorm:"column:withdraw_notice"`             // 提现充值管理员:true通知,false不通知
	WarnPct                float64 `gorm:"column:warn_pct"`                    // 警戒线:亏损达到该比例则触发警告线 0:不启用
	LowWarnCanBuy          bool    `gorm:"column:low_warn_can_buy"`            // 低于警戒线是否允许开仓:1允许开仓 2禁止
	ClosePct               float64 `gorm:"column:close_pct"`                   // 平仓线:亏损达到该比例则触发平仓线 0:不启用
	HolidayCharge          bool    `gorm:"column:holiday_charge"`              // 非交易日留仓收取管理费:1收取 2不收取
	BuyFee                 float64 `gorm:"column:buy_fee"`                     // 买入手续费率
	SellFee                float64 `gorm:"column:sell_fee"`                    // 卖出手续费率
	RegistCode             bool    `gorm:"column:regist_code"`                 // 是否启用推荐码(启用则必须要输入正确推荐码才能注册成功):1启用(是) 2不启用(否)
	BankNo                 string  `gorm:"column:bank_no"`                     // 收款行银行卡号
	BankName               string  `gorm:"column:bank_name"`                   // 收款人姓名
	BankAddr               string  `gorm:"column:bank_addr"`                   // 收款人开户行地址
	BankChannel            bool    `gorm:"column:bank_channel"`                // 银行卡支付账户:true 开启 2关闭
	QrcodeChannel          bool    `gorm:"column:qrcode_channel"`              // 支付宝支付通道:true开启 2关闭
	AlipayChannel          bool    `gorm:"column:alipay_channel"`              // 支付宝唤醒支付 true:开启 2关闭
	ContractLever          string  `gorm:"column:contract_lever"`              // 合约倍数:1,2,3,4,5,6,7,8,10
	IsSupportDayContract   bool    `gorm:"column:is_support_day_contract"`     // 是否支持按日合约
	IsSupportWeekContract  bool    `gorm:"column:is_support_week_contract"`    // 是否支持按周合约
	IsSupportMonthContract bool    `gorm:"column:is_support_month_contract"`   // 是否支持按月合约
	DayContractFee         float64 `gorm:"column:day_contract_fee"`            // 日管理费率
	WeekContractFee        float64 `gorm:"column:week_contract_fee"`           // 周管理费率
	MonthContractFee       float64 `gorm:"column:month_contract_fee"`          // 月管理费率
	MiniChargeFee          float64 `gorm:"column:mini_charge_fee"`             // 最低交易手续费:0不生效
	IsSupportBroker        bool    `gorm:"column:is_support_broker"`           // 是否对接券商
	AdminPhone             string  `json:"admin_phone"`                        // 管理员手机号码
}

///////////////////////////////////sysParam表///////////////////////////////////

const (
	SmsStatusValid   = 1 // 短信状态生效
	SmsStatusInValid = 2 // 短信状态失效
)

// Sms 短信验证码
type Sms struct {
	Phone  string    `gorm:"column:phone"`
	Code   string    `gorm:"column:code"`
	Status int8      `gorm:"column:status"` // '状态:1生效 2失效'
	Msg    string    `gorm:"column:msg"`
	Time   time.Time `gorm:"column:send_time"`
}

///////////////////////////////////sms表///////////////////////////////////

///////////////////////////////////funds表///////////////////////////////////

// Funds 资金表
type Funds struct {
	UID       int64   `gorm:"column:uid"`       // 用户ID
	Margin    float64 `gorm:"column:margin"`    // 保证金
	Available float64 `gorm:"column:available"` // 可用资金
	Freeze    float64 `gorm:"column:available"` // 冻结资金
}

///////////////////////////////////funds表///////////////////////////////////

///////////////////////////////////portfolio自选股表///////////////////////////////////

// Portfolio 自选股表
type Portfolio struct {
	UID        int64     `gorm:"column:uid"`        // 用户ID
	StockCode  string    `gorm:"column:code"`       // 股票代码
	StockName  string    `gorm:"column:name"`       // 股票名称
	CreateTime time.Time `gorm:"column:created_at"` // 创建时间
}

///////////////////////////////////portfolio自选股表///////////////////////////////////

///////////////////////////////////buy买入成交表///////////////////////////////////

// Buy 买入成交表
type Buy struct {
	ID          int64     `gorm:"column:id"`           // 主键ID
	EntrustID   int64     `gorm:"column:entrust_id"`   // 委托表ID
	UID         int64     `gorm:"column:uid"`          // 用户ID
	ContractID  int64     `gorm:"column:contract_id"`  // 合约编号
	OrderTime   time.Time `gorm:"column:order_time"`   // 订单时间
	StockCode   string    `gorm:"column:stock_code"`   // 股票代码
	StockName   string    `gorm:"column:stock_name"`   // 股票名称
	Price       float64   `gorm:"column:price"`        // 价格
	Amount      int64     `gorm:"column:amount"`       // 数量
	Balance     float64   `gorm:"column:balance"`      // 成交金额
	EntrustProp int64     `gorm:"column:entrust_prop"` // 委托类型:1限价 2市价
	Fee         float64   `gorm:"column:fee"`          // 交易手续费
	PositionID  int64     `gorm:"column:position_id"`  // 持仓表序号
}

///////////////////////////////////buy买入成交表///////////////////////////////////

///////////////////////////////////sell卖出成交表///////////////////////////////////

type Sell struct {
	ID            int64     `gorm:"column:id"`             // 主键ID
	EntrustID     int64     `gorm:"column:entrust_id"`     // 委托表ID
	UID           int64     `gorm:"column:uid"`            // 用户ID
	ContractID    int64     `gorm:"column:contract_id"`    // 合约编号
	OrderTime     time.Time `gorm:"column:order_time"`     // 订单时间
	StockCode     string    `gorm:"column:stock_code"`     // 股票代码
	StockName     string    `gorm:"column:stock_name"`     // 股票名称
	Price         float64   `gorm:"column:price"`          // 价格
	Amount        int64     `gorm:"column:amount"`         // 数量
	Balance       float64   `gorm:"column:balance"`        // 成交金额
	PositionPrice float64   `gorm:"column:position_price"` // 持仓价格
	Profit        float64   `gorm:"column:profit"`         // 盈亏金额
	EntrustProp   int64     `gorm:"column:entrust_prop"`   // 委托类型:1限价 2市价
	Fee           float64   `gorm:"column:fee"`            // 交易手续费
	PositionID    int64     `gorm:"column:position_id"`    // 持仓表序号
	Mode          int64     `gorm:"column:mode"`           // 类型:1 主动卖出 2系统平仓
	Reason        string    `gorm:"column:reason"`         // 系统平仓原因
}

///////////////////////////////////sell卖出成交表///////////////////////////////////

///////////////////////////////////msg消息表///////////////////////////////////

type Msg struct {
	ID         int64     `gorm:"column:id"`        // 主键ID
	UID        int64     `gorm:"column:uid"`       // 用户ID
	Title      string    `gorm:"column:title"`     // 标题
	Content    string    `gorm:"column:content"`   // 内容
	CreateTime time.Time `gorm:"column:create_at"` // 创建时间
}

///////////////////////////////////msg消息表///////////////////////////////////

///////////////////////////////////log消息表///////////////////////////////////

type Log struct {
	ID         int64     `gorm:"column:id"`         // 主键ID
	UID        int64     `gorm:"column:uid"`        // 用户ID
	Title      string    `gorm:"column:title"`      // 标题
	Content    string    `gorm:"column:content"`    // 内容
	Status     int64     `gorm:"column:status"`     // 状态:0 未读 1已读
	CreateTime time.Time `gorm:"column:created_at"` // 创建时间
}

///////////////////////////////////log消息表///////////////////////////////////

///////////////////////////////////TradeCalendar交易日历///////////////////////////////////

type TradeCalendar struct {
	Date  time.Time `gorm:"column:date"`  // 日期时间
	Trade bool      `gorm:"column:trade"` // 类型:true交易日 false非交易日
}

///////////////////////////////////TradeDate交易日表///////////////////////////////////
