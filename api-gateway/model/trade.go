package model

// InitTrade 接口返回结构
type InitTrade struct {
	ContractName   string          `json:"contract_name"`
	ContractID     int64           `json:"contract_id"`
	TotalAssets    float64         `json:"total_assets"` // 总市值
	ValMoney       float64         `json:"val_money"`
	Margin         float64         `json:"margin"`
	TotalProfit    float64         `json:"total_profit"`
	TodayProfit    float64         `json:"today_profit"`
	TodayProfitPct float64         `json:"today_profit_pct"` // 今日盈亏比例
	Positions      []*PositionItem `json:"list"`
}

// PositionItem 持仓item
type PositionItem struct {
	PositionID  int64   `json:"position_id"`  // 持仓编号
	StockCode   string  `json:"stock_code"`   // 股票代码
	StockName   string  `json:"stock_name"`   // 股票名称
	Amount      int64   `json:"amount"`       // 持仓数量
	DealPrice   float64 `json:"deal_price"`   // 成本价
	Profit      float64 `json:"profit"`       // 累计盈亏
	MarketValue float64 `json:"market_value"` // 市值
	ValAmount   int64   `json:"val_amount"`   // 可用股数
	NowPrice    float64 `json:"now_price"`    // 现价
	ProfitPct   float64 `json:"profit_pct"`   // 盈亏比率
}

// PositionDetail 持仓明细
type PositionDetail struct {
	TotalProfit        float64               `json:"total_profit"`     // 总盈亏
	TotalProfitPct     float64               `json:"total_profit_pct"` // 总盈亏比率
	BeginDate          string                `json:"begin_date"`       // 持仓开始日期
	EndDate            string                `json:"end_date"`         // 持仓结束日期
	MarketValue        float64               `json:"market_value"`     // 持仓市值(当前市值)
	InvestMoney        float64               `json:"invest_money"`     // 投入资金
	RetrieveMoney      float64               `json:"retrieve_money"`   // 回收资金
	TotalFee           float64               `json:"total_fee"`        // 交易手续费
	PositionDetailItem []*PositionDetailItem `json:"list"`             // 持仓明细item
}

// PositionDetailItem 持仓明细item
type PositionDetailItem struct {
	Date   string  `json:"date"`   // 日期
	Type   string  `json:"type"`   // 买入
	Amount int64   `json:"amount"` // 数量
	Price  float64 `json:"price"`  // 价格
	Money  float64 `json:"money"`  // 交易金额
	Fee    float64 `json:"fee"`    // 费用
}

// TradeDeal 成交
type TradeDeal struct {
	EntrustID int64   `json:"id"`         // 委托编号
	StockCode string  `json:"stock_code"` // 股票代码
	StockName string  `json:"stock_name"` // 股票名称
	Time      string  `json:"time"`       // 时间
	Type      int64   `json:"type"`       // 类型 1:买入 2卖出
	Price     float64 `json:"price"`      // 价格
	Amount    int64   `json:"amount"`     // 数量
	Balance   float64 `json:"balance"`    // 成交金额
}

// Fee 合约费用
type Fee struct {
	Name   string  `json:"name"`   // 名称(股票名称 & 合约)
	Date   string  `json:"date"`   // 日期
	Amount int64   `json:"amount"` // 数量
	Fee    float64 `json:"fee"`    // 费用
	Type   string  `json:"type"`   // 业务类型
}

// TradeDetail 成交明细
type TradeDetail struct {
	StockCode  string  `json:"stock_code"`  // 股票代码
	StockName  string  `json:"stock_name"`  // 股票名称
	Price      float64 `json:"price"`       // 价格
	Amount     int64   `json:"amount"`      // 数量
	Balance    float64 `json:"balance"`     // 成交金额
	Fee        float64 `json:"fee"`         // 交易手续费
	Status     string  `json:"status"`      // 交易状态
	Date       string  `json:"date"`        // 交易日期
	Time       string  `json:"time"`        // 交易时间
	EntrustID  int64   `json:"entrust_id"`  // 交易序号
	Type       string  `json:"type"`        // 交易类型
	ContractID int64   `json:"contract_id"` // 合约账户
}

// PositionResp 持仓界面返回
type PositionResp struct {
	ContractID   int64           `json:"contract_id"`
	ContractName string          `json:"contract_name"`
	ValMoney     float64         `json:"val_money"`
	Positions    []*PositionItem `json:"list"`
}

// InitBuy 买入界面初始化
type InitBuy struct {
	ContractID   int64   `json:"contract_id"`
	ContractName string  `json:"contract_name"`
	MaxBuyAmount int64   `json:"max_buy_amount"` // 最大可买数量
	PanKou       *PanKou `json:"pan_kou"`        // 盘口信息
}

// PanKou 盘口信息
type PanKou struct {
	StockCode      string  `json:"stock_code"`
	StockName      string  `json:"stock_name"`
	CurrentPrice   float64 `json:"current_price"`
	Chg            float64 `json:"chg"`              // 涨跌幅度
	ChgPct         float64 `json:"chg_pct"`          // 涨跌幅度百分比
	LimitUpPrice   float64 `json:"limit_up_price"`   // 涨停价
	LimitDownPrice float64 `json:"limit_down_price"` // 跌停价
	BuyPrice1      float64 `json:"buy_price_1"`
	BuyPrice2      float64 `json:"buy_price_2"`
	BuyPrice3      float64 `json:"buy_price_3"`
	BuyPrice4      float64 `json:"buy_price_4"`
	BuyPrice5      float64 `json:"buy_price_5"`
	BuyVol1        int64   `json:"buy_vol_1"`
	BuyVol2        int64   `json:"buy_vol_2"`
	BuyVol3        int64   `json:"buy_vol_3"`
	BuyVol4        int64   `json:"buy_vol_4"`
	BuyVol5        int64   `json:"buy_vol_5"`
	SellPrice1     float64 `json:"sell_price_1"`
	SellPrice2     float64 `json:"sell_price_2"`
	SellPrice3     float64 `json:"sell_price_3"`
	SellPrice4     float64 `json:"sell_price_4"`
	SellPrice5     float64 `json:"sell_price_5"`
	SellVol1       int64   `json:"sell_vol_1"`
	SellVol2       int64   `json:"sell_vol_2"`
	SellVol3       int64   `json:"sell_vol_3"`
	SellVol4       int64   `json:"sell_vol_4"`
	SellVol5       int64   `json:"sell_vol_5"`
}

// InitSell 卖出初始化
type InitSell struct {
	ContractID    int64   `json:"contract_id"`
	ContractName  string  `json:"contract_name"`
	MaxSellAmount int64   `json:"max_sell_amount"` // 最大可卖数量
	PanKou        *PanKou `json:"pan_kou"`         // 盘口信息
}

// Withdraw 撤单
type Withdraw struct {
	EntrustID     int64   `json:"entrust_id"`    // 委托id
	StockCode     string  `json:"stock_code"`    // 股票代码
	StockName     string  `json:"stock_name"`    // 股票名称
	Time          string  `json:"time"`          // 时间
	Type          int64   `json:"type"`          // 类型:1买入 2卖出
	EntrustPrice  float64 `json:"entrust_price"` // 委托价格
	DealPrice     float64 `json:"deal_price"`    // 成交价格
	EntrustAmount int64   `json:"amount"`        // 委托数量
	DealAmount    int64   `json:"deal_amount"`   // 成交数量
	DealStatus    bool    `json:"status"`        // 可撤单:true,不可撤单false
	StatusDesc    string  `json:"status_desc"`   // 未成交，全部成交,已撤单
}

const (
	UserMode   = 0 // 用户模式
	SystemMode = 1 // 系统模式
)

// EntrustPackage 交易委托pack
type EntrustPackage struct {
	UID         int64   // 用户UID
	ContractID  int64   // 合约ID
	Code        string  // 股票代码
	Price       float64 // 股票价格
	Amount      int64   // 股票数量
	EntrustProp int64   // 委托类型:1限价 2市价
	Mode        int64   // 类型:0 主动卖出 1系统平仓
}

// SellOut 清仓股票
type SellOut struct {
	PositionID int64   `json:"id"`         // 委托id
	StockCode  string  `json:"stock_code"` // 股票代码
	StockName  string  `json:"stock_name"` // 股票名称
	Time       string  `json:"time"`       // 时间
	Profit     float64 `json:"profit"`     // 盈亏金额
	ProfitPct  string  `json:"profit_pct"` // 盈亏比例
	Price      float64 `json:"price"`      // 清仓价格
}

// ConvertPanKou 盘口数据
func ConvertPanKou(qt *TencentQuote) *PanKou {
	return &PanKou{
		StockCode:      qt.Code,
		StockName:      qt.Name,
		CurrentPrice:   qt.CurrentPrice,
		Chg:            qt.Chg,
		ChgPct:         qt.ChgPercent,     // 涨跌幅度百分比
		LimitUpPrice:   qt.LimitUpPrice,   // 涨停价
		LimitDownPrice: qt.LimitDownPrice, // 跌停价
		BuyPrice1:      qt.BuyPrice1,
		BuyPrice2:      qt.BuyPrice2,
		BuyPrice3:      qt.BuyPrice3,
		BuyPrice4:      qt.BuyPrice4,
		BuyPrice5:      qt.BuyPrice5,
		BuyVol1:        qt.BuyVol1,
		BuyVol2:        qt.BuyVol2,
		BuyVol3:        qt.BuyVol3,
		BuyVol4:        qt.BuyVol4,
		BuyVol5:        qt.BuyVol5,
		SellPrice1:     qt.SellPrice1,
		SellPrice2:     qt.SellPrice2,
		SellPrice3:     qt.SellPrice3,
		SellPrice4:     qt.SellPrice4,
		SellPrice5:     qt.SellPrice5,
		SellVol1:       qt.SellVol1,
		SellVol2:       qt.SellVol2,
		SellVol3:       qt.SellVol3,
		SellVol4:       qt.SellVol4,
		SellVol5:       qt.SellVol5,
	}
}
