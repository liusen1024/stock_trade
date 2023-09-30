package model

type RoleModule struct {
	ID       int64  `gorm:"column:id" json:"id"`
	RoleID   int64  `gorm:"column:role_id" json:"role_id"`
	Module   string `gorm:"column:module" json:"module"`
	ModuleID int64  `gorm:"column:module_id" json:"module_id"`
}

const (
	UserPage     = 1 // 用户管理
	TradePage    = 2 // 股票交易
	ContractPage = 3 // 合约管理
	BrokerPage   = 4 // 券商管理
	StockPage    = 5 // 股票管理
	SystemPage   = 6 // 系统管理
	AgentPage    = 7 // 代理机构管理
	LogPage      = 8 // 日志管理
)

var RoleModuleMap = map[int64]string{
	UserPage:     "用户管理",
	TradePage:    "股票交易",
	ContractPage: "合约管理",
	BrokerPage:   "券商管理",
	StockPage:    "股票管理",
	SystemPage:   "系统管理",
	AgentPage:    "代理机构管理",
	LogPage:      "日志管理",
}
