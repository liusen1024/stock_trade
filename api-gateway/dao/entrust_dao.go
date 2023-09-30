package dao

import (
	"context"
	"gorm.io/gorm/clause"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"

	"gorm.io/gorm"
)

// EntrustDao 委托
type EntrustDao struct {
}

var _entrustDao = &EntrustDao{}

// EntrustDaoInstance 提供一个可用的对象
func EntrustDaoInstance() *EntrustDao {
	return _entrustDao
}

// GetTodayEntrust 查询今日委托
func (s *EntrustDao) GetTodayEntrust(ctx context.Context, contractID int64) ([]*model.Entrust, error) {
	var list []*model.Entrust
	sql := "select * from entrust where contract_id = ? and date(order_time) = CURRENT_DATE"
	err := db.StockDB().WithContext(ctx).Raw(sql, contractID).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeContractNoFound, "系统错误:查询持仓失败")
	}
	return list, nil
}

// GetAllEntrusts 查询所有的委托记录
func (s *EntrustDao) GetAllEntrusts(ctx context.Context, contractID int64) ([]*model.Entrust, error) {
	var list []*model.Entrust
	sql := "select * from entrust where contract_id = ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, contractID).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeContractNoFound, "系统错误:查询持仓失败")
	}
	return list, nil
}

// GetEntrustByTimeRange 查询所有的委托记录
func (s *EntrustDao) GetEntrustByTimeRange(ctx context.Context, contractID int64, beginDate, endDate string) ([]*model.Entrust, error) {
	var list []*model.Entrust
	sql := "select * from entrust where contract_id = ? and date(order_time) >= ? and date(order_time) <= ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, contractID, beginDate, endDate).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeContractNoFound, "系统错误:查询持仓失败")
	}
	return list, nil
}

// GetEntrustByContractID 根据contractID查询委托记录
func (s *EntrustDao) GetEntrustByContractID(ctx context.Context, contractID int64) ([]*model.Entrust, error) {
	var list []*model.Entrust
	sql := "select * from entrust where contract_id = ?"
	err := db.StockDB().WithContext(ctx).Raw(sql, contractID).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeContractNoFound, "系统错误:查询持仓失败")
	}
	return list, nil
}

// GetEntrustByID 根据ID查询委托记录
func (s *EntrustDao) GetEntrustByID(ctx context.Context, entrustID int64) (*model.Entrust, error) {
	var entrust *model.Entrust
	sql := "select * from entrust where id = ?"
	err := db.StockDB().WithContext(ctx).Raw(sql, entrustID).Take(&entrust).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, serr.New(serr.ErrCodeBusinessFail, "委托记录不存在")
		}
		return nil, serr.New(serr.ErrCodeContractNoFound, "系统错误:查询委托失败")
	}
	return entrust, nil
}

// CreateWithTx 创建委托:事务类型
func (s *EntrustDao) CreateWithTx(tx *gorm.DB, entrust *model.Entrust) (*model.Entrust, error) {
	if err := tx.Table("entrust").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&entrust).Error; err != nil {
		log.Errorf("填写委托表失败:%+v", err)
		return nil, err
	}
	return entrust, nil
}

// Update 更新委托表
func (s *EntrustDao) Update(ctx context.Context, entrust *model.Entrust) error {
	updateMap := make(map[string]interface{})
	updateMap["amount"] = entrust.Amount
	updateMap["price"] = entrust.Price
	updateMap["balance"] = entrust.Balance
	updateMap["status"] = entrust.Status
	updateMap["position_id"] = entrust.PositionID
	updateMap["fee"] = entrust.Fee
	updateMap["is_broker_entrust"] = entrust.IsBrokerEntrust
	updateMap["remark"] = entrust.Remark
	if err := db.StockDB().WithContext(ctx).Table("entrust").Where("id = ? ", entrust.ID).Updates(updateMap).Error; err != nil {
		log.Errorf("更新委托表失败:%+v", err)
		return err
	}
	return nil
}

// UpdateWithTx 更新委托表
func (s *EntrustDao) UpdateWithTx(tx *gorm.DB, entrust *model.Entrust) error {
	updateMap := make(map[string]interface{})
	updateMap["amount"] = entrust.Amount
	updateMap["price"] = entrust.Price
	updateMap["balance"] = entrust.Balance
	updateMap["status"] = entrust.Status
	updateMap["position_id"] = entrust.PositionID
	updateMap["deal_amount"] = entrust.DealAmount
	updateMap["fee"] = entrust.Fee
	updateMap["is_broker_entrust"] = entrust.IsBrokerEntrust
	if err := tx.Table("entrust").Where("id = ?", entrust.ID).Updates(updateMap).Error; err != nil {
		log.Errorf("更新委托表失败:%+v", err)
		return err
	}
	return nil
}

// Create 创建委托
func (s *EntrustDao) Create(ctx context.Context, entrust *model.Entrust) (*model.Entrust, error) {
	if err := db.StockDB().WithContext(ctx).Table("entrust").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&entrust).Error; err != nil {
		log.Errorf("填写委托表失败:%+v", err)
		return nil, err
	}
	return entrust, nil
}

// GetTodayEntrusts 查询当日的委托
func (s *EntrustDao) GetTodayEntrusts(ctx context.Context) ([]*model.Entrust, error) {
	var list []*model.Entrust
	sql := "select * from entrust where date(order_time) = CURRENT_DATE"
	err := db.StockDB().WithContext(ctx).Raw(sql).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeContractNoFound, "系统错误:查询委托表失败")
	}
	return list, nil
}

// GetEntrustsByIDs 根据ID查询委托记录
func (s *EntrustDao) GetEntrustsByIDs(ctx context.Context, entrustIDs []int64) ([]*model.Entrust, error) {
	var list []*model.Entrust
	if err := db.StockDB().WithContext(ctx).Table("entrust").Where("id in (?)", entrustIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpdateStatusWithTx 事务:更新状态
func (s *EntrustDao) UpdateStatusWithTx(tx *gorm.DB, entrust *model.Entrust) error {
	sql := "update entrust set status = ? where id = ? and status != ?"
	if err := tx.Exec(sql, entrust.Status, entrust.ID, entrust.Status).Error; err != nil {
		return err
	}
	return nil
}

// GetFirstBuyEntrustByPositionID 根据entrust的持仓ID查询第一次买入的委托
func (s *EntrustDao) GetFirstBuyEntrustByPositionID(ctx context.Context, entrust *model.Entrust) (*model.Entrust, error) {
	var res *model.Entrust
	sql := "select * from entrust where position_id = ? and entrust_bs = ? and status = 2"
	if err := db.StockDB().WithContext(ctx).Raw(sql, entrust.PositionID, model.EntrustBsTypeBuy).Take(&res).Error; err != nil {
		log.Errorf("查询失败:%+v,position_id:%+v entrust_bs:%+v", err, entrust.PositionID, model.EntrustBsTypeBuy)
		return nil, err
	}
	return res, nil
}
