package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"

	"gorm.io/gorm"
)

type ContractRecordDao struct{}

var _contractRecordDao = &ContractRecordDao{}

// ContractRecordDaoInstance 提供一个可用的对象
func ContractRecordDaoInstance() *ContractRecordDao {
	return _contractRecordDao
}

// GetContractByID 根据合约id查询合约
func (s *ContractRecordDao) GetContractByID(ctx context.Context, contractID int64) (*model.Contract, error) {
	var contract *model.Contract
	sql := "select * from contract_record where id = ? "
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

func (s *ContractRecordDao) Copy(ctx context.Context) error {
	if err := db.StockDB().WithContext(ctx).Exec("delete from contract_record").Error; err != nil {
		return err
	}
	if err := db.StockDB().WithContext(ctx).Exec("insert into contract_record (select * from contract where status=2)").Error; err != nil {
		return err
	}
	return nil
}
