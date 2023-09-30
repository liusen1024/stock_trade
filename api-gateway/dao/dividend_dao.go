package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"
)

// DividendDao 分红派息
type DividendDao struct{}

var _dividendDao = &DividendDao{}

// DividendDaoInstance 提供一个可用的对象
func DividendDaoInstance() *DividendDao {
	return _dividendDao
}

// GetDividendByPositionIDs 根据持仓id查询分红表
func (s *DividendDao) GetDividendByPositionIDs(ctx context.Context, positionID int64) ([]*model.Dividend, error) {
	var list []*model.Dividend
	sql := "select * from dividend where position_id = ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, positionID).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeBusinessFail, "系统错误:查询分红派息订单错误")
	}
	return list, nil
}

// GetByContractID 根据合约ID查询分红
func (s *DividendDao) GetByContractID(ctx context.Context, contractID int64) ([]*model.Dividend, error) {
	var list []*model.Dividend
	if err := db.StockDB().WithContext(ctx).Table("dividend").Where("contract_id = ?", contractID).Find(&list).Error; err != nil {
		log.Errorf("GetByContractID err:%+v", err)
		return nil, err
	}
	return list, nil
}

// Create 创建分红信息
func (s *DividendDao) Create(ctx context.Context, dividend *model.Dividend) error {
	if err := db.StockDB().WithContext(ctx).Table("dividend").Create(&dividend).Error; err != nil {
		log.Errorf("dividend create err:%+v", err)
		return err
	}
	return nil
}
