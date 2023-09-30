package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"

	"gorm.io/gorm/clause"
)

// PortfolioDao 短信
type PortfolioDao struct{}

var _portfolioDao = &PortfolioDao{}

// PortfolioDaoInstance 提供一个可用的对象
func PortfolioDaoInstance() *PortfolioDao {
	return _portfolioDao
}

// GetPortfolioList 根据uid查询自选股
func (s *PortfolioDao) GetPortfolioList(ctx context.Context, uid int64) ([]*model.Portfolio, error) {
	var list []*model.Portfolio
	sql := "select * from portfolio where uid = ? "
	err := db.StockDB().WithContext(ctx).Raw(sql, uid).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeBusinessFail, "系统错误:查询自选股失败")
	}
	return list, nil
}

// DeletePortfolio 删除自选股
func (s *PortfolioDao) DeletePortfolio(ctx context.Context, uid int64, code string) error {
	sql := "delete from portfolio where uid = ? and code = ?"
	if err := db.StockDB().WithContext(ctx).Exec(sql, uid, code).Error; err != nil {
		return err
	}
	return nil
}

// CreatePortfolio 删除自选股
func (s *PortfolioDao) CreatePortfolio(ctx context.Context, portfolio *model.Portfolio) error {
	if err := db.StockDB().WithContext(ctx).Table("portfolio").Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "code"}},
			UpdateAll: true,
		}).Create(&portfolio).Error; err != nil {
		return err
	}
	return nil
}
