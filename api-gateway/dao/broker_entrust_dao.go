package dao

import (
	"context"
	"gorm.io/gorm"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/common/log"

	"gorm.io/gorm/clause"
)

type BrokerEntrustDao struct{}

var _brokerEntrustDao = &BrokerEntrustDao{}

// BrokerEntrustDaoInstance 提供一个可用的对象
func BrokerEntrustDaoInstance() *BrokerEntrustDao {
	return _brokerEntrustDao
}

// Create 创建
func (s *BrokerEntrustDao) Create(ctx context.Context, entrust *model.BrokerEntrust) (*model.BrokerEntrust, error) {
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Create(&entrust).Error; err != nil {
		log.Errorf("Create err:%+v", err)
		return nil, err
	}
	return entrust, nil
}

// MCreate 批量创建券商委托表
func (s *BrokerEntrustDao) MCreate(ctx context.Context, list []*model.BrokerEntrust) error {
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&list).Error; err != nil {
		log.Errorf("MCreate err:%+v", err)
		return err
	}
	return nil
}

// MCreateWithTx 事务:批量创建券商委托表
func (s *BrokerEntrustDao) MCreateWithTx(tx *gorm.DB, list []*model.BrokerEntrust) error {
	if err := tx.Table("broker_entrust").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&list).Error; err != nil {
		log.Errorf("MCreate err:%+v", err)
		return err
	}
	return nil
}

// Update 更新
func (s *BrokerEntrustDao) Update(ctx context.Context, entrust *model.BrokerEntrust) error {
	updateMap := make(map[string]interface{})
	updateMap["entrust_amount"] = entrust.EntrustAmount   // 委托总数量
	updateMap["entrust_price"] = entrust.EntrustPrice     // 委托价格
	updateMap["entrust_balance"] = entrust.EntrustBalance // 委托总金额
	updateMap["deal_amount"] = entrust.DealAmount         // 成交数量
	updateMap["deal_price"] = entrust.DealPrice           // 成交价格
	updateMap["deal_balance"] = entrust.DealBalance       // 成交总金额
	updateMap["status"] = entrust.Status                  // 订单状态
	updateMap["fee"] = entrust.Fee                        // 券商交易总手续费
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Where("id = ? ", entrust.ID).Updates(updateMap).Error; err != nil {
		log.Errorf("更新券商委托表失败:%+v", err)
		return err
	}
	return nil
}

func (s *BrokerEntrustDao) UpdateByEntrustNo(ctx context.Context, entrust *model.BrokerEntrust) error {
	updateMap := make(map[string]interface{})
	updateMap["uid"] = entrust.UID
	updateMap["contract_id"] = entrust.ContractID
	updateMap["broker_id"] = entrust.BrokerID
	updateMap["order_time"] = entrust.OrderTime
	updateMap["stock_code"] = entrust.StockCode
	updateMap["stock_name"] = entrust.StockName
	updateMap["entrust_amount"] = entrust.EntrustAmount
	updateMap["entrust_price"] = entrust.EntrustPrice
	updateMap["entrust_balance"] = entrust.EntrustBalance
	updateMap["deal_amount"] = entrust.DealAmount
	updateMap["deal_price"] = entrust.DealPrice
	updateMap["status"] = entrust.Status
	updateMap["entrust_bs"] = entrust.EntrustBs
	updateMap["entrust_prop"] = entrust.EntrustProp
	updateMap["fee"] = entrust.Fee
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Where("broker_entrust_no = ?", entrust.BrokerEntrustNo).Updates(updateMap).Error; err != nil {
		log.Errorf("更新券商委托表失败:%+v", err)
		return err
	}
	return nil
}

// GetByIDs 根据id查询
func (s *BrokerEntrustDao) GetByIDs(ctx context.Context, ids []int64) ([]*model.BrokerEntrust, error) {
	var list []*model.BrokerEntrust
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Where("id in (?)", ids).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetByID 根据ID查询
func (s *BrokerEntrustDao) GetByID(ctx context.Context, id int64) (*model.BrokerEntrust, error) {
	var entrust *model.BrokerEntrust
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Where("id = ?", id).Take(&entrust).Error; err != nil {
		return nil, err
	}
	return entrust, nil
}

// GetByEntrustIDs 根据id查询
func (s *BrokerEntrustDao) GetByEntrustIDs(ctx context.Context, entrustIDs []int64) ([]*model.BrokerEntrust, error) {
	var list []*model.BrokerEntrust
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Where("entrust_id in (?)", entrustIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetByEntrustID 根据委托ID查询券商委托表,会返回券商委托列表
func (s *BrokerEntrustDao) GetByEntrustID(ctx context.Context, entrustID int64) ([]*model.BrokerEntrust, error) {
	var list []*model.BrokerEntrust
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Where("entrust_id = ?", entrustID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetTodayEntrusts 查询今日委托
func (s *BrokerEntrustDao) GetTodayEntrusts(ctx context.Context) ([]*model.BrokerEntrust, error) {
	var list []*model.BrokerEntrust
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Where("date(order_time)=CURRENT_DATE").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetTodayEntrustsByEntrustNos 根据委托编号查询委托
func (s *BrokerEntrustDao) GetTodayEntrustsByEntrustNos(ctx context.Context, entrustNos []string) ([]*model.BrokerEntrust, error) {
	var list []*model.BrokerEntrust
	if err := db.StockDB().WithContext(ctx).Table("broker_entrust").Where("broker_entrust_no in (?) and date(order_time)=CURRENT_DATE", entrustNos).Find(&list).Error; err != nil {
		log.Errorf("GetTodayEntrustsByEntrustNos err:%+v", err)
		return nil, err
	}
	return list, nil
}

//
//// CreateWithTx 事务:填写买入表
//func (s *BrokerEntrustDao) CreateWithTx(tx *gorm.DB, position *model.BrokerPosition) (*model.BrokerPosition, error) {
//	if err := tx.Table("broker_position").Clauses(clause.OnConflict{
//		Columns:   []clause.Column{{Name: "contract_id"}, {Name: "broker_id"}, {Name: "stock_code"}},
//		UpdateAll: true,
//	}).Create(&position).Error; err != nil {
//		log.Errorf("填写券商持仓表失败:%+v", err)
//		return nil, err
//	}
//	return position, nil
//}

//// GetBrokerPosition 根据合约ID、券商ID、股票代码,查询用户是否有券商持仓记录
//func (s *BrokerEntrustDao) GetBrokerPosition(ctx context.Context, contractID int64, brokerID int64, stockCode string) (*model.BrokerPosition, error) {
//	sql := "select * from broker_position where contract_id = ? and broker_id = ? stock_code = ?"
//	var position *model.BrokerPosition
//	if err := db.StockDB().WithContext(ctx).Raw(sql, contractID, brokerID, stockCode).Take(&position).Error; err != nil {
//		if err == gorm.ErrRecordNotFound {
//			return &model.BrokerPosition{}, nil
//		}
//		return nil, err
//	}
//	return position, nil
//}
