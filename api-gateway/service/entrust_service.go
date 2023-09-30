package service

import (
	"sync"
)

// EntrustService 委托服务
type EntrustService struct {
}

var (
	entrustService *EntrustService
	entrustOnce    sync.Once
)

// EntrustServiceInstance EntrustServiceInstance实例
func EntrustServiceInstance() *EntrustService {
	entrustOnce.Do(func() {
		entrustService = &EntrustService{}
	})
	return entrustService
}
