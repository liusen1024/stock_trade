package model

import "time"

const (
	TransferTypeRecharge       = 1 // 充值
	TransferTypeWithdraw       = 2 // 提现
	TransferTypeAppendMoney    = 3 // 追加保证金
	TransferTypeExpandMoney    = 4 // 扩大合约
	TransferTypeCloseContract  = 5 // 终止合约
	TransferTypeGetMoney       = 6 // 合约提盈
	TransferTypeCreateContract = 7 // 创建合约
	TransferStatusPre          = 0 // 预插入
	TransferStatusWaitExam     = 1 // 待审核
	TransferStatusSuccess      = 2 // 成功
	TransferStatusFail         = 3 // 失败
)

// Transfer 银行转账表
type Transfer struct {
	ID        int64     `gorm:"column:id"`         // 主键ID
	UID       int64     `gorm:"column:uid"`        // 用户ID
	OrderTime time.Time `gorm:"column:order_time"` // 订单时间
	Money     float64   `gorm:"column:money"`      // 金额
	Type      int64     `gorm:"column:type"`       // 类型：1充值 2提现
	Status    int64     `gorm:"column:status"`     // 状态:0预插入 1待审核 2成功 3失败
	Name      string    `gorm:"column:name"`       // 提现收款人
	BankNo    string    `gorm:"column:bank_no"`    // 提现银行卡号
	Channel   string    `gorm:"column:channel"`    // 渠道
	OrderNo   string    `gorm:"column:order_no"`   // 订单号流水
}

func (t *Transfer) TransferConvertTitle() string {
	title := ""
	switch t.Type {
	case TransferTypeRecharge:
		title = "转入"
	case TransferTypeWithdraw:
		title = "转出"
	case TransferTypeAppendMoney:
		title = "合约追加资金"
	case TransferTypeExpandMoney:
		title = "合约扩大资金"
	case TransferTypeCloseContract:
		title = "合约终止提取保证金"
	case TransferTypeGetMoney:
		title = "合约提盈"
	case TransferTypeCreateContract:
		title = "创建合约"
	}

	switch t.Status {
	case TransferStatusPre:
		title += "预插入"
	case TransferStatusWaitExam:
		title += "待审核"
	case TransferStatusSuccess:
		title += "成功"
	case TransferStatusFail:
		title += "失败"
	}
	return title
}
