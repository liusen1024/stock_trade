package id_gen

import (
	"fmt"
	"sync"
	"time"

	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake
var so sync.Once
var startTime = time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)

func init() {
	setupSF()
}

func setupSF() {
	so.Do(func() {
		// 使用默认的低16位IP作为机器ID
		settings := sonyflake.Settings{
			StartTime: startTime,
		}
		sf = sonyflake.NewSonyflake(settings)
	})
}

// GetNextID 获取唯一自增 ID
func GetNextID() int64 {
	if sf == nil {
		setupSF()
		if sf == nil {
			panic("failed to init sony flake")
		}
	}
	id, err := sf.NextID()
	if err != nil {
		panic(fmt.Errorf("sony flake over 174 years err: %+v", err))
	}
	result := int64(id)
	return result
}
