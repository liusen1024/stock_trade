package dao

import (
	"context"
	"gorm.io/gorm/clause"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/common/log"
)

type BrokerErrorLogDao struct{}

var _brokerErrorLogDao = &BrokerErrorLogDao{}

// BrokerErrorLogDaoInstance 提供一个可用的对象
func BrokerErrorLogDaoInstance() *BrokerErrorLogDao {
	return _brokerErrorLogDao
}

// Create 创建券商错误日志
func (s *BrokerErrorLogDao) Create(ctx context.Context, errorLog *model.BrokerErrorLog) error {
	if err := db.StockDB().WithContext(ctx).Table("broker_error_log").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&errorLog).Error; err != nil {
		log.Errorf("create err:%+v", err)
		return err
	}
	return nil
}
