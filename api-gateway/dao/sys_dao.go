package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"
	"time"
)

// SysDao 系统
type SysDao struct{}

var _sysDao = &SysDao{}

// SysDaoInstance 提供一个可用的对象
func SysDaoInstance() *SysDao {
	return _sysDao
}

func sysCacheKey() string {
	return "sys_conf"
}

// GetSysParam 系统参数
func (s *SysDao) GetSysParam(ctx context.Context) (*model.SysParam, error) {
	var sys *model.SysParam
	err := db.GetOrLoad(ctx, sysCacheKey(), 24*time.Hour, &sys, func() error {
		sql := "select * from sysparam limit 1"
		err := db.StockDB().WithContext(ctx).Raw(sql).Take(&sys).Error
		if err != nil {
			return serr.New(serr.ErrCodeBusinessFail, "系统参数错误")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sys, nil
}

// Update 更新系统参数
func (s *SysDao) Update(ctx context.Context, sys *model.SysParam) error {
	sql := "delete from sysparam"
	if err := db.StockDB().WithContext(ctx).Exec(sql).Error; err != nil {
		return err
	}
	if err := db.StockDB().WithContext(ctx).Table("sysparam").Create(sys).Error; err != nil {
		log.Errorf("更新系统参数失败:%+v", err)
		return err
	}
	if err := db.RedisClient().Del(ctx, sysCacheKey()).Err(); err != nil {
		log.Errorf("删除缓存失败")
		return err
	}
	return nil
}
