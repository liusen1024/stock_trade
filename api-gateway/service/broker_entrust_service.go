package service

import (
	"sync"
)

// BrokerEntrustService 委托服务
type BrokerEntrustService struct {
}

var (
	brokerEntrustService *BrokerEntrustService
	brokerEntrustOnce    sync.Once
)

// BrokerEntrustServiceInstance 实例
func BrokerEntrustServiceInstance() *BrokerEntrustService {
	brokerEntrustOnce.Do(func() {
		brokerEntrustService = &BrokerEntrustService{}
	})
	return brokerEntrustService
}
