package dao

import (
	"context"
	"errors"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContractDao struct{}

var _contractDao = &ContractDao{}

// ContractDaoInstance 提供一个可用的对象
func ContractDaoInstance() *ContractDao {
	return _contractDao
}

const contractTable = "contract"

// UpdateContract 更新合约
func (s *ContractDao) UpdateContract(ctx context.Context, contract *model.Contract) error {
	err := db.StockDB().WithContext(ctx).Table(contractTable).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&contract).Error
	if err != nil {
		log.Errorf("UpdateContract err:%+v", err)
		return err
	}
	return nil
}

// UpdateWithTx 更新合约
func (s *ContractDao) UpdateWithTx(tx *gorm.DB, contract *model.Contract) error {
	if err := tx.Table(contractTable).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&contract).Error; err != nil {
		log.Errorf("UpdateContract err:%+v", err)
		return err
	}
	return nil
}

// CreateContract 创建合约
func (s *ContractDao) CreateContract(ctx context.Context, contract *model.Contract) (*model.Contract, error) {
	err := db.StockDB().WithContext(ctx).Table("contract").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&contract).Error
	if err != nil {
		log.Errorf("创建合约失败:%+v", contract)
		return nil, err
	}
	return contract, nil
}

// SetContractStatus 设置合约状态
func (s *ContractDao) SetContractStatus(ctx context.Context, contractID, status int64) error {
	sql := "update contract set status = ? where id = ?"
	err := db.StockDB().WithContext(ctx).Exec(sql, status, contractID).Error
	if err != nil {
		log.Errorf("更新资金表失败:id[%v] status[%v]", contractID, status)
		return serr.New(serr.ErrCodeBusinessFail, "设置合约状态失败")
	}
	return nil
}

// GetContractByID 根据合约id查询合约
func (s *ContractDao) GetContractByID(ctx context.Context, contractID int64) (*model.Contract, error) {
	var contract *model.Contract
	sql := "select * from contract where id = ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, contractID).Take(&contract).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, serr.New(serr.ErrCodeBusinessFail, "合约不存在")
		}
		log.Errorf("GetContractByID err:%+v", err)
		return nil, err
	}
	return contract, nil
}

// GetContractByIDWithTx 根据合约id查询合约
func (s *ContractDao) GetContractByIDWithTx(tx *gorm.DB, contractID int64) (*model.Contract, error) {
	var contract *model.Contract
	sql := "select * from contract where id = ? "
	if err := tx.Raw(sql, contractID).Take(&contract).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("合约不存在")
		}
		return nil, err
	}
	return contract, nil
}

// GetEnableContractByUID 查找uid一个有效的contract
func (s *ContractDao) GetEnableContractByUID(ctx context.Context, uid int64) (*model.Contract, error) {
	var list []*model.Contract
	sql := "select * from contract where uid = ? and status = ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, uid, model.ContractStatusEnable).Find(&list).Error
	if err != nil {
		log.Errorf("查询生效合约错误:%+v", err)
		return nil, err
	}
	for _, it := range list {
		if it.Status == model.ContractStatusEnable {
			return it, nil
		}
	}
	return nil, serr.New(serr.ErrCodeContractNoFound, "请申请合约")
}

// UpdateContractValMoney 更新合约可用资金
func (s *ContractDao) UpdateContractValMoney(ctx context.Context, valMoney float64, contractID int64) error {
	sql := "update contract set val_money = ? where id = ?"
	if err := db.StockDB().WithContext(ctx).Exec(sql, valMoney, contractID).Error; err != nil {
		log.Errorf("刷新可用资金失败:%+v", err)
		return err
	}
	return nil
}

// GetContractsByUID 根据uid查询合约
func (s *ContractDao) GetContractsByUID(ctx context.Context, uid int64) ([]*model.Contract, error) {
	var list []*model.Contract
	if err := db.StockDB().WithContext(ctx).Table("contract").Where("uid = ?", uid).Find(&list).Error; err != nil {
		log.Errorf("GetContractsByUID err:%+v", err)
		return nil, err
	}
	return list, nil
}

func (s *ContractDao) GetContracts(ctx context.Context) ([]*model.Contract, error) {
	var list []*model.Contract
	if err := db.StockDB().WithContext(ctx).Table("contract").Find(&list).Error; err != nil {
		log.Errorf("GetContracts err:%+v", err)
		return nil, err
	}
	return list, nil
}
