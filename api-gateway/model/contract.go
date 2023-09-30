package model

import (
	"fmt"
	"stock/api-gateway/util"
	"time"
)

// ContractConf 合约申请页面初始化
type ContractConf struct {
	Money float64          `json:"money"` // 可用资金
	Type  []*ContractType  `json:"type"`  // 类型
	Lever []*ContractLever `json:"lever"` // 合约杠杆
}

// 合约类型
const (
	ContractTypeDay   = 1
	ContractTypeWeek  = 2
	ContractTypeMonth = 3
)

type ContractType struct {
	Type int64  `json:"type"` // 合约类型
	Name string `json:"name"` // 类型名称
}

type ContractLever struct {
	Lever int64  `json:"lever"` //	杠杆系数
	Name  string `json:"name"`  // 杠杆名称
}

// ContractTypeMap 合约类型映射
var ContractTypeMap = map[int64]*ContractType{
	ContractTypeDay: {
		Type: ContractTypeDay, // 合约类型
		Name: "按天合约",          // 类型名称
	},
	ContractTypeWeek: {
		Type: ContractTypeWeek, // 合约类型
		Name: "按周合约",           // 类型名称
	},
	ContractTypeMonth: {
		Type: ContractTypeMonth, // 合约类型
		Name: "按月合约",            // 类型名称
	},
}

///////////////////////////////////Contract合约表///////////////////////////////////

// ContractRiskLevel 合约风险类型
type ContractRiskLevel int64
type ContractWithdrawStatus int64

const (
	ContractStatusApply    = 1 // 1预申请
	ContractStatusEnable   = 2 // 2操盘中
	ContractStatusDisabled = 3 // 3操盘结束

	ContractRiskLevelHealth ContractRiskLevel = 1 // 合约安全状态
	ContractRiskLevelWarn   ContractRiskLevel = 2 // 触发警戒线
	ContractRiskLevelClose  ContractRiskLevel = 3 // 触发平仓线

	ContractWithdrawStatusEnable  ContractWithdrawStatus = 1 // 可撤单
	ContractWithdrawStatusDisable ContractWithdrawStatus = 2 // 不可撤单
)

type Contract struct {
	ID           int64     `gorm:"column:id"`            // 主键ID
	UID          int64     `gorm:"column:uid"`           // 用户ID
	InitMoney    float64   `gorm:"column:init_money"`    // 原始保证金
	Money        float64   `gorm:"column:money"`         // 现保证金
	ValMoney     float64   `gorm:"column:val_money"`     // 可用资金
	Lever        int64     `gorm:"column:lever"`         // 合约杠杠倍数
	Status       int64     `gorm:"column:status"`        // 合约状态:1预申请 2操盘中 3操盘结束
	Type         int64     `gorm:"column:type"`          // 合约类型:1按天合约 2:按周合约 3:按月合约
	AppendMoney  float64   `gorm:"column:append_money"`  // 追加金额
	OrderTime    time.Time `gorm:"column:order_time"`    // 订单时间
	CloseTime    time.Time `gorm:"column:close_time"`    // 关闭时间
	CloseExplain string    `gorm:"column:close_explain"` // 关闭说明
}

// Balance 合约原始总资金
func (c *Contract) Balance() float64 {
	return c.InitMoney * float64(c.Lever)
}

// FullName 合约全称:按日10倍合约
func (c *Contract) FullName() string {
	return fmt.Sprintf("按%s%d倍合约", c.TypeText(), c.Lever)
}

func (c *Contract) TypeText() string {
	switch c.Type {
	case ContractTypeDay:
		return "日"
	case ContractTypeWeek:
		return "周"
	case ContractTypeMonth:
		return "月"
	}
	return ""
}

///////////////////////////////////Contract合约表///////////////////////////////////

// ContractApply 合约提交返回信息
type ContractApply struct {
	ContractName string  `json:"contract_name"` // 合约名称
	Money        float64 `json:"money"`         // 投资本金
	ContractID   int64   `json:"contract_id"`   // 合约id
	ContractType string  `json:"contract_type"` // 合约类型
	ValMoney     float64 `json:"val_money"`     // 操盘资金
	Period       string  `json:"period"`        // 操盘期限
	Interest     string  `json:"interest"`      // 利息
	Warn         string  `json:"warn"`          // 警戒线
	Close        string  `json:"close"`         // 平仓线
	Pay          float64 `json:"pay"`           // 支付本金
	Wallet       float64 `json:"wallet"`        // 钱包余额
}

// ContractDetail 合约详情
type ContractDetail struct {
	ID                    int64   `json:"id"`                      // 合约id
	Name                  string  `json:"name"`                    // 合约名称
	Profit                float64 `json:"profit"`                  // 持仓盈亏
	ProfitPct             float64 `json:"profit_pct"`              // 持仓盈亏比率
	MarketValue           float64 `json:"market_value"`            // 证券市值
	Money                 float64 `json:"money"`                   // 保证金
	ValMoney              float64 `json:"val_money"`               // 可用资金
	InterestBearingAmount float64 `json:"interest_bearing_amount"` // 计息金额
	Interest              string  `json:"interest"`                // 利息
	AppendMoney           float64 `json:"append_money"`            // 追加保证金
	Warn                  float64 `json:"warn"`                    // 警戒参考值线
	Close                 float64 `json:"close"`                   // 平仓线参考值
	Risk                  int64   `json:"risk"`                    // 风险水平
	CreateTime            string  `json:"create_time"`             // 合约创建时间
	TotalAsset            string  `json:"total_asset"`             // 总资产
	OriginalMoney         string  `json:"original_money"`          // 原始保证金
	Lever                 string  `json:"lever"`                   // 合约杠杠
	GetProfit             string  `json:"get_profit"`              // 可提取利润
}

// AppendExpandContract 追加扩大合约
type AppendExpandContract struct {
	ID    int64   `json:"id"`    // 合约ID
	Name  string  `json:"name"`  // 合约名称
	Money float64 `json:"money"` // 可追加的保证金
}

// ValidContract 有效合约
type ValidContract struct {
	ID          int64   `json:"id"`           // 合约id
	Name        string  `json:"name"`         // 合约名称
	TodayProfit float64 `json:"today_profit"` // 今日盈亏
	Profit      float64 `json:"profit"`       // 持仓盈亏
	ProfitPct   float64 `json:"profit_pct"`   // 持仓盈亏比率
	Money       float64 `json:"money"`        // 保证金
	MarketValue float64 `json:"market_value"` // 证券市值
	ValMoney    float64 `json:"val_money"`    // 可用资金
	Interest    string  `json:"interest"`     // 管理费
	Warn        float64 `json:"warn"`         // 警戒参考值线
	Close       float64 `json:"close"`        // 平仓线参考值
	Risk        int64   `json:"risk"`         // 风险水平
	Select      bool    `json:"select"`       // 当前合约(true为选中)
}

// CalculateTodayProfit 计算合约的今日盈亏
func CalculateTodayProfit(position, yesterdayPosition []*Position, entrusts []*Entrust, qts map[string]*TencentQuote) float64 {
	if len(position) == 0 {
		return 0.00
	}
	var todayProfit float64
	// 查询昨日收盘价格

	// 当前价格*当前持仓数量
	for _, it := range position {
		todayProfit += qts[it.StockCode].CurrentPrice * float64(it.Amount)
	}
	// 昨日收盘价*昨日持仓数量
	for _, it := range yesterdayPosition {
		todayProfit -= qts[it.StockCode].ClosePrice * float64(it.Amount)
	}
	// + 当日卖出金额（含费） && -当日买入金额（含费）

	for _, it := range entrusts {
		if it.Status == EntrustStatusTypeDeal && it.EntrustBS == EntrustBsTypeBuy {
			todayProfit -= it.Balance - it.Fee
		} else if it.Status == EntrustStatusTypeDeal && it.EntrustBS == EntrustBsTypeSell {
			todayProfit += it.Balance - it.Fee
		}
	}
	return todayProfit
}

// Interest 合约利息
func Interest(contract *Contract, sys *SysParam, money float64) float64 {
	// 计算利息
	interest := 0.00
	switch contract.Type {
	case ContractTypeDay:
		interest = sys.DayContractFee
	case ContractTypeWeek:
		interest = sys.WeekContractFee
	case ContractTypeMonth:
		interest = sys.MonthContractFee
	}
	return util.FloatRound(money*float64(contract.Lever)*interest, 2)
}

// HisContract 历史合约
type HisContract struct {
	ID           int64   `json:"id"`            // 合约id
	Name         string  `json:"name"`          // 合约名称
	InvestMoney  float64 `json:"money"`         // 投入资金
	RecoverMoney float64 `json:"recover_money"` // 收回资金
	Profit       float64 `json:"profit"`        // 盈亏金额
	CreateTime   string  `json:"create_time"`   // 创建时间
	CloseTime    string  `json:"close_time"`    // 结算时间
}

// GetContract 查询合约
type GetContract struct {
	ID       int64   `json:"id"`        // 合约id
	Name     string  `json:"name"`      // 合约名称
	Money    float64 `json:"money"`     // 保证金
	ValMoney float64 `json:"val_money"` // 可用资金
	Select   bool    `json:"select"`    // 当前选中合约
}
