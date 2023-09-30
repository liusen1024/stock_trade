package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
)

type RoleModuleDao struct {
}

var _roleModuleDao = &RoleModuleDao{}

// RoleModuleDaoInstance 提供一个可用的对象
func RoleModuleDaoInstance() *RoleModuleDao {
	return _roleModuleDao
}

// GetModules getModules
func (s *RoleModuleDao) GetModules(ctx context.Context) ([]*model.RoleModule, error) {
	var list []*model.RoleModule
	if err := db.StockDB().WithContext(ctx).Table("role_module").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetModulesByRoleID GetModulesByRoleID
func (s *RoleModuleDao) GetModulesByRoleID(ctx context.Context, roleID int64) ([]*model.RoleModule, error) {
	var list []*model.RoleModule
	if err := db.StockDB().WithContext(ctx).Table("role_module").Where("role_id = ?", roleID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *RoleModuleDao) Delete(ctx context.Context, roleID int64) error {
	sql := "delete from role_module where role_id = ?"
	if err := db.StockDB().WithContext(ctx).Exec(sql, roleID).Error; err != nil {
		return err
	}
	return nil
}

func (s *RoleModuleDao) Create(ctx context.Context, modules []*model.RoleModule) error {
	if err := db.StockDB().WithContext(ctx).Table("role_module").Create(&modules).Error; err != nil {
		return err
	}
	return nil
}
