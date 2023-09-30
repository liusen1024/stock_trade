package model

type DownloadInterface interface {
}

type UserListResp struct {
	ID             int64   `json:"id"`
	UserName       string  `json:"user_name"`
	Password       string  `json:"password"`
	Name           string  `json:"name"`
	Agent          string  `json:"agent"`
	Money          float64 `json:"money"`
	FreezeMoney    float64 `json:"freeze_money"`
	Broker         bool    `json:"broker"`
	Contract       int64   `json:"contract"`
	Authentication bool    `json:"authentication"`
	Online         bool    `json:"online"`
	RegisterTime   string  `json:"register_time"`
	Status         bool    `json:"status"`
}

type RechargeResp struct {
	ID       int64   `json:"id"`
	UserName string  `json:"user_name"`
	Name     string  `json:"name"`
	Agent    string  `json:"agent"`
	Time     string  `json:"time"`
	Money    float64 `json:"money"`
	OrderNo  string  `json:"order_no"`
	Channel  string  `json:"channel"`
	Status   int64   `json:"status"` // 1待审核 2成功 3失败
}

type WithdrawResp struct {
	ID       int64   `json:"id"`
	UserName string  `json:"user_name"`
	Name     string  `json:"name"`
	Agent    string  `json:"agent"`
	Time     string  `json:"time"`
	Money    float64 `json:"money"`
	BankName string  `json:"bank_name"`
	BankNo   string  `json:"bank_no"`
	Status   int64   `json:"status"` // 1待审核 2成功 3失败
}

type TradeBuyResp struct {
	ID            int64   `json:"id"` // ID->委托序号
	UserName      string  `json:"user_name"`
	Name          string  `json:"name"`
	Agent         string  `json:"agent"`
	Time          string  `json:"time"`
	ContractID    int64   `json:"contract_id"`
	ContractName  string  `json:"contract_name"`
	StockCode     string  `json:"stock_code"`
	StockName     string  `json:"stock_name"`
	Price         float64 `json:"price"`
	Amount        int64   `json:"amount"`
	Balance       float64 `json:"balance"`
	Profit        float64 `json:"profit"`
	ProfitPct     string  `json:"profit_pct"`
	Type          string  `json:"type"`
	Fee           float64 `json:"fee"`
	Broker        bool    `json:"broker"`
	BrokerAccount string  `json:"broker_account"`
	BrokerOrderNo string  `json:"broker_order_no"`
}

type TradeSellResp struct {
	ID            int64   `json:"id"` // ID->委托序号
	UserName      string  `json:"user_name"`
	Name          string  `json:"name"`
	Agent         string  `json:"agent"`
	Time          string  `json:"time"`
	ContractID    int64   `json:"contract_id"`
	ContractName  string  `json:"contract_name"`
	StockCode     string  `json:"stock_code"`
	StockName     string  `json:"stock_name"`
	Price         float64 `json:"price"`
	Amount        int64   `json:"amount"`
	Balance       float64 `json:"balance"`
	Type          string  `json:"type"`
	Fee           float64 `json:"fee"`
	Broker        bool    `json:"broker"`
	BrokerAccount string  `json:"broker_account"`
	BrokerOrderNo string  `json:"broker_order_no"`
}

type TradePositionResp struct {
	ID            int64   `json:"id"` // ID->委托序号
	UserName      string  `json:"user_name"`
	Name          string  `json:"name"`
	Agent         string  `json:"agent"`
	Time          string  `json:"time"`
	ContractID    int64   `json:"contract_id"`
	ContractName  string  `json:"contract_name"`
	StockCode     string  `json:"stock_code"`
	StockName     string  `json:"stock_name"`
	PositionPrice float64 `json:"position_price"`
	CurrentPrice  float64 `json:"current_price"`
	Amount        int64   `json:"amount"`
	FreezeAmount  int64   `json:"freeze_amount"`
	Profit        float64 `json:"profit"`
	ProfitPct     string  `json:"profit_pct"`
	Broker        bool    `json:"broker"`
}

type TradeDividendResp struct {
	ID             int64   `json:"id"`
	UserName       string  `json:"user_name"`
	Name           string  `json:"name"`
	Agent          string  `json:"agent"`
	Time           string  `json:"time"`
	ContractID     int64   `json:"contract_id"`
	ContractName   string  `json:"contract_name"`
	StockCode      string  `json:"stock_code"`
	StockName      string  `json:"stock_name"`
	PositionPrice  float64 `json:"position_price"`  // 持仓价格
	PositionAmount int64   `json:"amount"`          // 持仓股数
	DividendAmount string  `json:"dividend_amount"` // 转股比例
	DividendMoney  string  `json:"dividend_money"`  // 现金分红比例
	IsBuyBack      bool    `json:"buy_back"`        // 是否零股回购
	BuyBackAmount  int64   `json:"buy_back_amount"` // 零股回购数量
	BuyBackPrice   float64 `json:"buy_back_price"`  // 零股回购价格
}

type TradeDetailResp struct {
	ID             int64             `json:"id"`
	UserName       string            `json:"user_name"`
	Name           string            `json:"name"`
	Agent          string            `json:"agent"`
	Time           string            `json:"time"`
	ContractID     int64             `json:"contract_id"`
	ContractName   string            `json:"contract_name"`
	Broker         bool              `json:"broker"`
	BrokerAccount  string            `json:"brokerAccount"`
	BuyDetailItem  []*BuyDetailItem  `json:"buy_list"`
	SellDetailItem []*SellDetailItem `json:"sell_list"`
}

type BuyDetailItem struct {
	ID        int64   `json:"id"`         // 买入交易序号
	OrderTime string  `json:"order_time"` // 交易时间
	StockCode string  `json:"stock_code"`
	StockName string  `json:"stock_name"`
	Price     float64 `json:"deal_price"`
	Amount    int64   `json:"amount"`
	Balance   float64 `json:"balance"`
	Type      string  `json:"type"`
	Fee       float64 `json:"fee"`
}

type SellDetailItem struct {
	ID            int64   `json:"id"`         // 买入交易序号
	OrderTime     string  `json:"order_time"` // 交易时间
	StockCode     string  `json:"stock_code"`
	StockName     string  `json:"stock_name"`
	PositionPrice float64 `json:"position_price"`
	DealPrice     float64 `json:"deal_price"`
	Amount        int64   `json:"amount"`
	Balance       float64 `json:"balance"` // 成交金额
	Profit        float64 `json:"profit"`  // 盈亏金额
	Fee           float64 `json:"fee"`     // 手续费
	Type          string  `json:"type"`    // 交易类型
	Mode          string  `json:"mode"`    // 交易方式
	Reason        string  `json:"reason"`  // 系统平仓原因
}

type TradeEntrustResp struct {
	ID            int64   `json:"id"`
	UserName      string  `json:"user_name"`
	Name          string  `json:"name"`
	Agent         string  `json:"agent"`
	Time          string  `json:"time"`
	ContractID    int64   `json:"contract_id"`
	ContractName  string  `json:"contract_name"`
	StockCode     string  `json:"stock_code"`
	StockName     string  `json:"stock_name"`
	Price         float64 `json:"price"`
	Amount        int64   `json:"amount"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`
	Broker        bool    `json:"broker"`
	BrokerAccount string  `json:"broker_account"`
	BrokerOrderNo string  `json:"broker_order_no"`
	Remark        string  `json:"remark"`
}

type CmsContractResp struct {
	ID           int64   `json:"id"`            // 主键ID
	ContractName string  `json:"contract_name"` // 合约名称
	UserName     string  `json:"user_name"`     // 用户名称
	Name         string  `json:"name"`          // 用户姓名
	Agent        string  `json:"agent"`         // 代理机构
	Time         string  `json:"time"`          // 时间
	Asset        float64 `json:"asset"`         // 总资产
	MarketValue  float64 `json:"market_value"`  // 持仓市值
	Profit       float64 `json:"profit"`        // 盈亏金额
	InitMoney    float64 `json:"init_money"`    // 原始保证金
	Money        float64 `json:"money"`         // 现保证金
	ValMoney     float64 `json:"val_money"`     // 可用资金
	AppendMoney  float64 `json:"append_money"`  // 追加保证金
	Warn         float64 `json:"warn"`          // 警戒线
	Close        float64 `json:"close"`         // 平仓线
	Status       string  `json:"status"`        // 合约状态
	Risk         string  `json:"risk"`          // 合约风控
}

type CmsContractFeeResp struct {
	ID           int64   `json:"id"`            // 主键ID
	ContractName string  `json:"contract_name"` // 合约名称
	UserName     string  `json:"user_name"`     // 用户名称
	Name         string  `json:"name"`          // 用户姓名
	Agent        string  `json:"agent"`         // 代理机构
	Time         string  `json:"time"`          // 时间
	StockCode    string  `json:"stock_code"`    // 股票代码
	StockName    string  `json:"stock_name"`    // 股票名称
	Balance      float64 `json:"balance"`       // 交易金额
	Item         string  `json:"item"`          // 费项
	Detail       string  `json:"detail"`        // 明细
}

type CmsBrokerResp struct {
	ID              int64   `form:"id" json:"id"`
	Priority        int64   `form:"priority" json:"priority"`
	IP              string  `form:"ip" json:"ip"`
	Port            int64   `form:"port" json:"port"`
	Name            string  `form:"name" json:"name"`
	Version         string  `form:"version" json:"version"`
	BranchNo        int64   `form:"branch_no" json:"branch_no"`
	Account         string  `form:"account" json:"account"`
	Password        string  `form:"password" json:"password"`
	CommPassword    string  `form:"comm_password" json:"comm_password"`
	SHHolderAccount string  `form:"sh_holder_account" json:"sh_holder_account"`
	SZHolderAccount string  `form:"sz_holder_account" json:"sz_holder_account"`
	Status          bool    `form:"status" json:"status"`
	Asset           float64 `json:"asset"` // 总资产
	ValMoney        float64 `json:"val_money"`
}

type CmsBrokerEntrustResp struct {
	ID         int64   `json:"id"`          // 券商编号
	BrokerName string  `json:"broker_name"` // 券商名称
	UserName   string  `json:"user_name"`   // 用户账户
	Name       string  `json:"name"`        // 用户姓名
	Agent      string  `json:"agent"`       // 代理机构
	Time       string  `json:"time"`        // 时间
	StockCode  string  `json:"stock_code"`  // 股票代码
	StockName  string  `json:"stock_name"`  // 股票名称
	Price      float64 `json:"price"`       // 委托价格
	Amount     int64   `json:"amount"`      // 委托数量
	DealAmount int64   `json:"deal_amount"` // 成交数量
	Status     string  `json:"status"`      // 状态
	Type       string  `json:"type"`        // 交易类型
	Prop       string  `json:"prop"`        // 委托类型
	EntrustNo  string  `json:"entrust_no"`  // 券商委托编号
}

// CmsBrokerPositionResp 券商管理-持仓
type CmsBrokerPositionResp struct {
	BrokerID      int64   `json:"id"`             // 券商编号
	BrokerName    string  `json:"name"`           // 券商名称
	StockCode     string  `json:"stock_code"`     // 股票代码
	StockName     string  `json:"stock_name"`     // 股票名称
	PositionPrice float64 `json:"position_price"` // 持仓价格
	CurrentPrice  float64 `json:"current_price"`  // 当前价格
	Amount        int64   `json:"amount"`         // 持仓数量
	FreezeAmount  int64   `json:"freeze_amount"`  // 冻结股数
}

type CmsStockDataResp struct {
	ID        int64  `json:"id"`
	StockCode string `json:"stock_code"`
	StockName string `json:"stock_name"`
	IPODate   string `json:"ipo_date"`
	Status    bool   `json:"status"`
}
