package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"

	"gorm.io/gorm"
)

// ContractFeeDao 委托
type ContractFeeDao struct {
}

var _contractFeeDao = &ContractFeeDao{}

// ContractFeeDaoInstance 提供一个可用的对象
func ContractFeeDaoInstance() *ContractFeeDao {
	return _contractFeeDao
}

// GetContractFeeByID 查询合约费用by合约id
func (s *ContractFeeDao) GetContractFeeByID(ctx context.Context, contractID int64) ([]*model.ContractFee, error) {
	var list []*model.ContractFee
	sql := "select * from contract_fee where contract_id = ?"
	err := db.StockDB().WithContext(ctx).Raw(sql, contractID).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeContractNoFound, "系统错误:查询失败")
	}
	return list, nil
}

// GetContractFeeByIDs 根据合约ID查询合约费用
func (s *ContractFeeDao) GetContractFeeByIDs(ctx context.Context, contractIDs []int64) ([]*model.ContractFee, error) {
	var list []*model.ContractFee
	if err := db.StockDB().WithContext(ctx).Table("contract_fee").Where("contract_id in (?)", contractIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CreateWithTx 手续费
func (s *ContractFeeDao) CreateWithTx(tx *gorm.DB, fee *model.ContractFee) error {
	if err := tx.Table("contract_fee").Create(&fee).Error; err != nil {
		log.Errorf("手续费扣除失败:%+v", err)
		return err
	}
	return nil
}

// Create 手续费
func (s *ContractFeeDao) Create(ctx context.Context, fee *model.ContractFee) error {
	if err := db.StockDB().WithContext(ctx).Table("contract_fee").Create(&fee).Error; err != nil {
		log.Errorf("手续费扣除失败:%+v", err)
		return err
	}
	return nil
}
