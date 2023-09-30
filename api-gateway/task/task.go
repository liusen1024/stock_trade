package service

import (
	"stock/common/log"
	"sync"

	"github.com/robfig/cron"
)

// TaskService 定时器列表管理
type TaskService struct {
	mainCron *cron.Cron
	schedule map[string]func()
}

var taskService *TaskService
var taskServiceOnce sync.Once

// TaskServiceInstance 单例，内部可以做内存cache
func TaskServiceInstance() *TaskService {
	taskServiceOnce.Do(func() {
		taskService = &TaskService{
			mainCron: cron.New(),
			schedule: map[string]func(){
				//"* * * * *":     testSecond, // test :每秒钟执行一次
				"0 50 23 * * ?": HisTrade,       // 每天23:50分执行一次 持仓归档为历史持仓
				"0 55 23 * * ?": ContractRecord, // 每天23:00分执行一次 归档本日合约
			},
		}
		taskService.loadTask()
	})
	return taskService
}

func (s *TaskService) loadTask() {
	for spec, fn := range s.schedule {
		s.mainCron.AddFunc(spec, fn)
	}
	s.mainCron.Start()
}

func testSecond() {
	log.Info("testSecond xxxx")
}
