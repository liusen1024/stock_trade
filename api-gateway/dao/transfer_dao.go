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

// TransferDao 转账
type TransferDao struct{}

var _transferDao = &TransferDao{}

// TransferDaoInstance 提供一个可用的对象
func TransferDaoInstance() *TransferDao {
	return _transferDao
}

func (s *TransferDao) CreateWithTx(tx *gorm.DB, transfer *model.Transfer) error {
	if err := tx.Table("transfer").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&transfer).Error; err != nil {
		log.Errorf("创建转账记录表失败:%+v", err)
		return err
	}
	return nil
}

// Create 创建
func (s *TransferDao) Create(ctx context.Context, transfer *model.Transfer) error {
	if err := db.StockDB().WithContext(ctx).Table("transfer").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&transfer).Error; err != nil {
		log.Errorf("创建转账记录表失败:%+v", err)
		return err
	}
	return nil
}

// GetByUid 根据用户ID查询消息
func (s *TransferDao) GetByUid(ctx context.Context, uid int64) ([]*model.Transfer, error) {
	var list []*model.Transfer
	err := db.StockDB().WithContext(ctx).Table("transfer").Where("uid = ?", uid).Find(&list).Error
	if err != nil {
		return nil, serr.ErrBusiness("查询失败")
	}
	return list, nil
}

// GetByOrderNo 根据订单号查询
func (s *TransferDao) GetByOrderNo(ctx context.Context, orderNo string) (*model.Transfer, error) {
	var transfer *model.Transfer
	if err := db.StockDB().WithContext(ctx).Table("transfer").Where("order_no = ?", orderNo).Take(&transfer).Error; err != nil {
		return nil, err
	}
	return transfer, nil
}

// GetByID 根据ID查询
func (s *TransferDao) GetByID(ctx context.Context, id int64) (*model.Transfer, error) {
	var transfer *model.Transfer
	if err := db.StockDB().WithContext(ctx).Table("transfer").Where("id = ?", id).Take(&transfer).Error; err != nil {
		return nil, err
	}
	return transfer, nil
}
