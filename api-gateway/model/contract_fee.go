package model

import "time"

const (
	ContractFeeDirectionPay    int64 = 1 // 费用方向:支出
	ContractFeeDirectionIncome int64 = 2 // 费用方向:收入

	ContractFeeTypeBuy         int64 = 1 // 买入费用
	ContractFeeTypeSell        int64 = 2 // 卖出费用
	ContractFeeTypeInterest    int64 = 3 // 合约利息费用
	ContractFeeTypeProfit      int64 = 4 // 卖出盈亏
	ContractFeeTypeAppendMoney int64 = 5 // 合约追加保证金
	ContractFeeTypeExpandMoney int64 = 6 // 合约扩大资金
	ContractFeeTypeClose       int64 = 7 // 合约结算资金
	ContractFeeTypeGetProfit   int64 = 8 // 合约提盈
)

var ContractFeeTypeMap = map[int64]string{
	ContractFeeTypeBuy:         "买入费用",
	ContractFeeTypeSell:        "卖出费用",
	ContractFeeTypeInterest:    "合约利息",
	ContractFeeTypeProfit:      "卖出盈亏",
	ContractFeeTypeAppendMoney: "追加保证金",
	ContractFeeTypeExpandMoney: "扩大资金",
	ContractFeeTypeClose:       "合约结算",
	ContractFeeTypeGetProfit:   "合约提盈",
}

func ContractFeeType(feeType string) int64 {
	for k, v := range ContractFeeTypeMap {
		if v == feeType {
			return k
		}
	}
	return 0
}

// ContractFee 合约费用
type ContractFee struct {
	ID         int64     `json:"id"`
	UID        int64     `json:"uid"`
	ContractID int64     `json:"contract_id"`
	Code       string    `json:"code"`       // 股票代码
	Name       string    `json:"name"`       // 股票名称
	Amount     int64     `json:"amount"`     // 股票交易数量
	OrderTime  time.Time `json:"order_time"` // 订单时间
	Direction  int64     `json:"direction"`  // 方向:1支出 2:收入
	Money      float64   `json:"money"`      // 金额
	Detail     string    `json:"detail"`     // 明细
	Type       int64     `json:"type"`       // 费用类型1:买入手续费 2:卖出手续费 3:合约利息 4:卖出盈亏 5:追加保证金 6:扩大资金 7:合约结算
}
