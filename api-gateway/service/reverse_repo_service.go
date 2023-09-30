package service

import (
	"sync"
	"time"
)

// ReverseRepoService 国债逆回购
type ReverseRepoService struct {
}

var (
	reverseRepoService *ReverseRepoService
	reverseRepoOnce    sync.Once
)

// ReverseRepoServiceInstance 实例
func ReverseRepoServiceInstance() *ReverseRepoService {
	reverseRepoOnce.Do(func() {
		reverseRepoService = &ReverseRepoService{}
		//ctx := context.Background()

		// 国债逆回购业务
		go func() {
			for range time.Tick(2 * time.Second) {

			}
		}()
	})
	return reverseRepoService
}
