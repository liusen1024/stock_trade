package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/common/log"

	"gorm.io/gorm/clause"
)

type TradeCalendarDao struct{}

var _tradeCalendarDao = &TradeCalendarDao{}

// TradeCalendarDaoInstance 提供一个可用的对象
func TradeCalendarDaoInstance() *TradeCalendarDao {
	return _tradeCalendarDao
}

// GetCalendar 查询交易日历
func (s *TradeCalendarDao) GetCalendar(ctx context.Context) ([]*model.TradeCalendar, error) {
	var list []*model.TradeCalendar
	if err := db.StockDB().WithContext(ctx).Table("trade_calendar").Find(&list).Error; err != nil {
		log.Errorf("查询GetCalendar数据库错误:%+v", err)
		return nil, err
	}
	return list, nil
}

// Create 创建交易日历
func (s *TradeCalendarDao) Create(ctx context.Context, calendar []*model.TradeCalendar) error {
	if err := db.StockDB().WithContext(ctx).Table("trade_calendar").Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "date"}},
			UpdateAll: true,
		}).Create(calendar).Error; err != nil {
		log.Errorf("插入交易日历失败:%+v", err)
		return err
	}
	return nil
}
