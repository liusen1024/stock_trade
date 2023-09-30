package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"

	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

type BrokerDao struct{}

var _brokerDao = &BrokerDao{}

// BrokerDaoInstance 提供一个可用的对象
func BrokerDaoInstance() *BrokerDao {
	return _brokerDao
}

// GetBroker 根据ID查询券商
func (s *BrokerDao) GetBroker(ctx context.Context, id int64) (*model.Broker, error) {
	var broker *model.Broker
	if err := db.StockDB().WithContext(ctx).Table("broker").Where("id = ?", id).Take(&broker).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, serr.New(serr.ErrCodeBusinessFail, "券商通道不存在")
		}
		log.Errorf("GetBroker err:%+v", err)
		return nil, err
	}
	return broker, nil
}

// GetBrokers 查询所有的券商
func (s *BrokerDao) GetBrokers(ctx context.Context) ([]*model.Broker, error) {
	var list []*model.Broker
	sql := "select * from broker"
	err := db.StockDB().WithContext(ctx).Raw(sql).Find(&list).Error
	if err != nil {
		log.Errorf("GetBrokers err:%+v", err)
		return nil, err
	}
	return list, nil
}

// GetBrokersByIDs 根据id查询券商
func (s *BrokerDao) GetBrokersByIDs(ctx context.Context, ids []int64) ([]*model.Broker, error) {
	var list []*model.Broker
	if err := db.StockDB().WithContext(ctx).Table("broker").Where("id in (?)", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *BrokerDao) Create(ctx context.Context, broker *model.Broker) error {
	if err := db.StockDB().WithContext(ctx).Table("broker").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&broker).Error; err != nil {
		log.Errorf("create err:%+v", err)
		return err
	}
	return nil
}
