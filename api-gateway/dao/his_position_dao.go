package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"
)

type HisPositionDao struct {
}

var _hisPositionDao = &HisPositionDao{}

// HisPositionDaoInstance 提供一个可用的对象
func HisPositionDaoInstance() *HisPositionDao {
	return _hisPositionDao
}

// GetYesterdayPositionByContractID 查询昨日持仓
func (s *HisPositionDao) GetYesterdayPositionByContractID(ctx context.Context, contractID int64) ([]*model.Position, error) {
	var list []*model.Position
	sql := "select * from his_position where contract_id = ? and date(order_time)=DATE_SUB(CURDATE(), INTERVAL 1 DAY)"
	err := db.StockDB().WithContext(ctx).Raw(sql, contractID).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeContractNoFound, "系统错误:查询历史持仓失败")
	}
	return list, nil
}

// GetYesterdayPositions 查询昨日持仓
func (s *HisPositionDao) GetYesterdayPositions(ctx context.Context) ([]*model.Position, error) {
	var list []*model.Position
	sql := "select * from his_position where date(order_time)=DATE_SUB(CURDATE(), INTERVAL 1 DAY)"
	if err := db.StockDB().WithContext(ctx).Raw(sql).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Create 创建历史持仓表
func (s *HisPositionDao) Create(ctx context.Context, positions []*model.Position) error {
	if err := db.StockDB().WithContext(ctx).Table("his_position").
		Create(&positions).Error; err != nil {
		log.Errorf("插入历史订单失败:%+v", err)
		return err
	}
	return nil
}

func (s *HisPositionDao) Del(ctx context.Context) error {
	sql := "delete from his_position"
	if err := db.StockDB().WithContext(ctx).Exec(sql).Error; err != nil {
		log.Errorf("Del err:%+v", err)
		return err
	}
	return nil
}
