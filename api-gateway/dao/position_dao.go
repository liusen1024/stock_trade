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

type PositionDao struct {
}

var _positionDao = &PositionDao{}

// PositionDaoInstance 提供一个可用的对象
func PositionDaoInstance() *PositionDao {
	return _positionDao
}

// GetPositionByContractID 根据合约查询持仓
func (s *PositionDao) GetPositionByContractID(ctx context.Context, contractID int64) ([]*model.Position, error) {
	var list []*model.Position
	sql := "select * from position where contract_id = ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, contractID).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeBusinessFail, "系统错误:查询持仓失败")
	}
	return list, nil
}

// GetPositionByEntrustID 通过持仓编号查询持仓记录
func (s *PositionDao) GetPositionByEntrustID(ctx context.Context, entrustID int64) (*model.Position, error) {
	var result *model.Position
	sql := "select * from position where entrust_id = ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, entrustID).Take(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, serr.New(serr.ErrCodeBusinessFail, "查询失败")
	}
	return result, nil
}

// GetPositions 查询所有持仓数据
func (s *PositionDao) GetPositions(ctx context.Context) ([]*model.Position, error) {
	var list []*model.Position
	sql := "select * from position"
	if err := db.StockDB().WithContext(ctx).Raw(sql).Find(&list).Error; err != nil {
		log.Errorf("查询持仓数据失败:%+v", err)
		return nil, err
	}
	return list, nil
}

// FreezePositionAmount 卖出:冻结持仓(事务)
//func (s *PositionDao) FreezePositionAmount(tx *gorm.DB, p *model.BindParam) error {
//	sql := "update position set freeze_amount = freeze_amount + ? where id = ?"
//	err := tx.Exec(sql, p.Amount, p.Position.ID).Error
//	if err != nil {
//		return serr.New(serr.ErrCodeBusinessFail, "委托失败:冻结股数失败")
//	}
//	return nil
//}

// CreateWithTx 事务:添加持仓表
func (s *PositionDao) CreateWithTx(tx *gorm.DB, position *model.Position) (*model.Position, error) {
	if err := tx.Table("position").Create(&position).Error; err != nil {
		log.Errorf("填写买入表失败:%+v", err)
		return nil, err
	}
	return position, nil
}

// UpdateWithTx 事务:更新持仓
func (s *PositionDao) UpdateWithTx(tx *gorm.DB, position *model.Position) error {
	updateMap := make(map[string]interface{})
	updateMap["order_time"] = position.OrderTime
	updateMap["price"] = position.Price
	updateMap["amount"] = position.Amount
	updateMap["freeze_amount"] = position.FreezeAmount
	updateMap["balance"] = position.Balance
	if err := tx.Table("position").Where("id = ? ", position.ID).Updates(updateMap).Error; err != nil {
		log.Errorf("持仓更新错误:%+v", err)
		return err
	}
	return nil
}

// Update 更新持仓
func (s *PositionDao) Update(ctx context.Context, position *model.Position) error {
	return db.StockDB().WithContext(ctx).Table("position").Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(position).Error
}

// DeleteWithTx 删除持仓记录
func (s *PositionDao) DeleteWithTx(tx *gorm.DB, position *model.Position) error {
	sql := "delete from position where id = ?"
	if err := tx.Exec(sql, position.ID).Error; err != nil {
		log.Errorf("删除持仓记录错误:%+v", err)
		return err
	}
	return nil
}

// UnFreezeAmount 解冻股票
func (s *PositionDao) UnFreezeAmount(ctx context.Context, contractID int64, code string, amount int64) error {
	sql := "update position set freeze_amount = freeze_amount - ? where contract_id = ? and stock_code = ?"
	if err := db.StockDB().WithContext(ctx).Exec(sql, amount, contractID, code).Error; err != nil {
		log.Errorf("解冻股票数量失败:%+v", err)
		return err
	}
	return nil
}

// UnFreezeAmountWithTx 事务:解冻股票
func (s *PositionDao) UnFreezeAmountWithTx(tx *gorm.DB, contractID int64, code string, amount int64) error {
	sql := "update position set freeze_amount = freeze_amount - ? where contract_id = ? and stock_code = ?"
	if err := tx.Exec(sql, amount, contractID, code).Error; err != nil {
		log.Errorf("解冻股票数量失败:%+v", err)
		return err
	}
	return nil
}

// GetContractPositionByCode 根据ContractID和Code查询持仓
func (s *PositionDao) GetContractPositionByCode(ctx context.Context, contractID int64, code string) (*model.Position, error) {
	var position *model.Position
	sql := "select * from position where contract_id = ? and stock_code = ?"
	if err := db.StockDB().WithContext(ctx).Raw(sql, contractID, code).Take(&position).Error; err != nil {
		return nil, err
	}
	return position, nil
}

// GetContractPositionByCodeWithTx 根据ContractID和Code查询持仓
func (s *PositionDao) GetContractPositionByCodeWithTx(tx *gorm.DB, contractID int64, code string) (*model.Position, error) {
	var position *model.Position
	sql := "select * from position where contract_id = ? and stock_code = ?"
	if err := tx.Raw(sql, contractID, code).Take(&position).Error; err != nil {
		return nil, err
	}
	return position, nil
}

// FreezeAmountWithTx 冻结股票
func (s *PositionDao) FreezeAmountWithTx(tx *gorm.DB, contractID int64, code string, amount int64) error {
	sql := "update position set freeze_amount = freeze_amount + ? where contract_id = ? and stock_code = ?"
	if err := tx.Exec(sql, amount, contractID, code).Error; err != nil {
		log.Errorf("冻结股票:%+v", err)
		return err
	}
	return nil
}

// GetPositionByID 根据ID查询持仓
func (s *PositionDao) GetPositionByID(ctx context.Context, positionID int64) (*model.Position, error) {
	var position *model.Position
	if err := db.StockDB().WithContext(ctx).Table("position").Where("id = ?", positionID).Take(&position).Error; err != nil {
		return nil, err
	}
	return position, nil
}
