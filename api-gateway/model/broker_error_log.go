package model

// BrokerErrorLog 券商错误日志
type BrokerErrorLog struct {
	ID    int64  `gorm:"column:id"`
	URL   string `gorm:"column:url"`
	Error string `gorm:"column:error"`
}
