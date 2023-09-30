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

type SellDao struct {
}

var _sellDao = &SellDao{}

// SellDaoInstance 提供一个可用的对象
func SellDaoInstance() *SellDao {
	return _sellDao
}

// GetByPositionIDs 根据持仓id查询卖出记录
func (s *SellDao) GetByPositionIDs(ctx context.Context, positionID []int64) ([]*model.Sell, error) {
	var list []*model.Sell
	sql := "select * from sell where position_id in ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, positionID).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeBusinessFail, "系统错误:查询卖出订单错误")
	}
	return list, nil
}

// GetByPositionID 根据持仓id查询卖出记录
func (s *SellDao) GetByPositionID(ctx context.Context, positionID int64) ([]*model.Sell, error) {
	var list []*model.Sell
	if err := db.StockDB().WithContext(ctx).Table("sell").Where("position_id = ?", positionID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CreateWithTx 事务:创建卖出记录
func (s *SellDao) CreateWithTx(tx *gorm.DB, sell *model.Sell) (*model.Sell, error) {
	if err := tx.Table("sell").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&sell).Error; err != nil {
		log.Errorf("填写卖出表失败:%+v", err)
		return nil, err
	}
	return sell, nil
}

// GetSellByEntrustIDs 根据委托id查询卖出记录
func (s *SellDao) GetSellByEntrustIDs(ctx context.Context, ids []int64) ([]*model.Sell, error) {
	var list []*model.Sell
	sql := "select * from sell where entrust_id in (?) "
	err := db.StockDB().WithContext(ctx).Raw(sql, ids).Find(&list).Error
	if err != nil {
		return nil, serr.ErrBusiness("查询订单失败")
	}
	return list, nil
}

// GetByEntrustID 根据委托ID查询卖出
func (s *SellDao) GetByEntrustID(ctx context.Context, entrustID int64) (*model.Sell, error) {
	var sell *model.Sell
	if err := db.StockDB().WithContext(ctx).Table("sell").Where("entrust_id = ?", entrustID).Take(&sell).Error; err != nil {
		return nil, err
	}
	return sell, nil
}

// GetByContractID 根据合约ID查询卖出记录
func (s *SellDao) GetByContractID(ctx context.Context, contractID int64) ([]*model.Sell, error) {
	var list []*model.Sell
	if err := db.StockDB().WithContext(ctx).Table("sell").Where("contract_id = ?", contractID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
