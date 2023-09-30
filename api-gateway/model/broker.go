package model

import "time"

type TDXQueryType int64

const (
	BrokerStatusEnable   = 1 // 1激活
	BrokerStatusDisabled = 2 // 2冻结

	TDXQueryTypeFund         TDXQueryType = 0 // 通达信查询资金
	TDXQueryTypePosition     TDXQueryType = 1 // 通达信查询持仓
	TDXQueryTypeTodayEntrust TDXQueryType = 2 // 当日委托
	TDXQueryTypeTodayDeal    TDXQueryType = 3 // 当日成交
	TDXQueryTypeWithdraw     TDXQueryType = 4 // 可撤单
)

type Broker struct {
	ClientID        int64             `gorm:"-"`                        // 券商连接成功id
	ID              int64             `gorm:"column:id"`                // 主键
	IP              string            `gorm:"column:ip"`                // 券商交易服务器IP地址
	Port            int64             `gorm:"column:port"`              // 券商交易服务器端口号
	Version         string            `gorm:"column:version"`           // 通达信客户端版本号
	BranchNo        int64             `gorm:"column:branch_no"`         // 营业部代码
	FundAccount     string            `gorm:"column:account"`           // 资金账号
	TradeAccount    string            `gorm:"column:trade_account"`     // 交易账号
	TradePassword   string            `gorm:"column:trade_password"`    // 交易密码
	TxPassword      string            `gorm:"column:tx_password"`       // 通讯密码
	SHHolderAccount string            `gorm:"column:sh_holder_account"` // 上海股东代码
	SZHolderAccount string            `gorm:"column:sz_holder_account"` // 深证股东代码
	Priority        int64             `gorm:"column:priority"`          // 顺序,数字越大,优先级越高
	Status          int64             `gorm:"column:status"`            // 状态:1激活 2冻结
	BrokerName      string            `gorm:"column:broker_name"`       // 券商名称
	ValMoney        float64           `gorm:"column:val_money"`         // 可用资金
	Asset           float64           `gorm:"column:asset"`             // 总资产
	CreateTime      time.Time         `gorm:"column:create_at"`         // 时间
	BrokerPosition  []*BrokerPosition `gorm:"-"`                        // 券商持仓
	IoTimes         int64             `gorm:"-"`                        // 超时次数
}

// BrokerPosition 券商持仓
type BrokerPosition struct {
	StockCode     string  `json:"stock_code"`     // 股票代码
	StockName     string  `json:"stock_name"`     // 股票名称
	Amount        int64   `json:"amount"`         // 总数量
	FreezeAmount  int64   `json:"freeze_amount"`  // 冻结数量
	PositionPrice float64 `json:"position_price"` // 持仓价格
	CurrentPrice  float64 `json:"current_price"`  // 当前价格
}

type BuyInfo struct {
	BrokerEntrustID string
}

type BrokerResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}
