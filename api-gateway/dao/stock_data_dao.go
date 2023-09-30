package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"

	"gorm.io/gorm/clause"
)

// StockDataDao 股票
type StockDataDao struct{}

var _stockDataDao = &StockDataDao{}

// StockDataDaoInstance 提供一个可用的对象
func StockDataDaoInstance() *StockDataDao {
	return _stockDataDao
}

// GetStockDataByCode 根据股票代码查询股票信息
func (s *StockDataDao) GetStockDataByCode(ctx context.Context, code string) (*model.StockData, error) {
	var result *model.StockData
	sql := "select * from stock_data where code = ?"
	err := db.StockDB().WithContext(ctx).Raw(sql, code).Take(&result).Error
	if err != nil {
		log.Errorf("查询stock_data表失败 err:%v", err)
		return nil, serr.New(serr.ErrCodeBusinessFail, "证券代码不存在")
	}
	return result, nil
}

func (s *StockDataDao) Get(ctx context.Context) ([]*model.StockData, error) {
	var list []*model.StockData
	if err := db.StockDB().WithContext(ctx).Table("stock_data").Find(&list).Error; err != nil {
		return nil, serr.New(serr.ErrCodeBusinessFail, "证券代码不存在")
	}
	return list, nil
}

// Update 更新股票列表
func (s *StockDataDao) Update(ctx context.Context, list []*model.StockData) error {
	if err := db.StockDB().WithContext(ctx).Table("stock_data").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(&list).Error; err != nil {
		log.Errorf("更新股票列表失败:%+v", err)
		return err
	}
	return nil
}

// UpdateStatusByID 根据ID查询更新
func (s *StockDataDao) UpdateStatusByID(ctx context.Context, id int64, status int64) error {
	sql := "update stock_data set status = ? where id = ?"
	if err := db.StockDB().WithContext(ctx).Exec(sql, status, id).Error; err != nil {
		return err
	}
	return nil
}
