package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BuyDao struct{}

var _buyDao = &BuyDao{}

// BuyDaoInstance 提供一个可用的对象
func BuyDaoInstance() *BuyDao {
	return _buyDao
}

func (s *BuyDao) GetBuyByPositionIDs(ctx context.Context, positionID []int64) ([]*model.Buy, error) {
	var list []*model.Buy
	sql := "select * from buy where position_id in ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, positionID).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeBusinessFail, "系统错误:查询买入订单错误")
	}
	return list, nil
}

// GetByPositionID 根据持仓ID查询买入记录
func (s *BuyDao) GetByPositionID(ctx context.Context, positionID int64) ([]*model.Buy, error) {
	var list []*model.Buy
	if err := db.StockDB().WithContext(ctx).Table("buy").Where("position_id = ?", positionID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CreateWithTx 事务:填写买入表
func (s *BuyDao) CreateWithTx(tx *gorm.DB, buy *model.Buy) error {
	if err := tx.Table("buy").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&buy).Error; err != nil {
		log.Errorf("填写买入表失败:%+v", err)
		return err
	}
	return nil
}

// GetByEntrustIDs 根据委托id查询买入记录
func (s *BuyDao) GetByEntrustIDs(ctx context.Context, entrustIDs []int64) ([]*model.Buy, error) {
	var list []*model.Buy
	sql := "select * from buy where entrust_id in (?)"
	if err := db.StockDB().WithContext(ctx).Raw(sql, entrustIDs).Find(&list).Error; err != nil {
		log.Errorf("GetBuyByEntrustIDs 错误:%+v", err)
		return nil, serr.ErrBusiness("查询记录失败")
	}
	return list, nil
}

// GetByEntrustID 根据委托id查询买入记录
func (s *BuyDao) GetByEntrustID(ctx context.Context, entrustID int64) (*model.Buy, error) {
	var buy *model.Buy
	sql := "select * from buy where entrust_id = ?"
	if err := db.StockDB().WithContext(ctx).Raw(sql, entrustID).Take(&buy).Error; err != nil {
		return nil, err
	}
	return buy, nil
}

// GetByContractID 根据合约ID查询买入记录
func (s *BuyDao) GetByContractID(ctx context.Context, contractID int64) ([]*model.Buy, error) {
	var list []*model.Buy
	if err := db.StockDB().WithContext(ctx).Table("buy").Where("contract_id = ?", contractID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
