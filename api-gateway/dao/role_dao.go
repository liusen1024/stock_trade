package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/common/log"

	"gorm.io/gorm/clause"
)

type RoleDao struct {
}

var _roleDao = &RoleDao{}

// RoleDaoInstance 提供一个可用的对象
func RoleDaoInstance() *RoleDao {
	return _roleDao
}

// GetRoles getRoles
func (s *RoleDao) GetRoles(ctx context.Context) ([]*model.Role, error) {
	var list []*model.Role
	sql := "select * from role "
	err := db.StockDB().WithContext(ctx).Raw(sql).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetRoleByUserName 根据用户名查询角色信息
func (s *RoleDao) GetRoleByUserName(ctx context.Context, userName string) (*model.Role, error) {
	var role *model.Role
	if err := db.StockDB().WithContext(ctx).Table("role").Where("user_name = ?", userName).Take(&role).Error; err != nil {
		return nil, err
	}
	return role, nil
}

// GetRolesByName GetRolesByName
func (s *RoleDao) GetRolesByName(ctx context.Context, roleName []string) ([]*model.Role, error) {
	var list []*model.Role
	sql := "select * from role where user_name in (?)"
	if err := db.StockDB().WithContext(ctx).Raw(sql, roleName).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *RoleDao) Create(ctx context.Context, role *model.Role) (*model.Role, error) {
	if err := db.StockDB().WithContext(ctx).Table("role").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&role).Error; err != nil {
		log.Errorf("create err:%+v", err)
		return nil, err
	}
	return role, nil
}

func (s *RoleDao) GetRoleByID(ctx context.Context, id int64) (*model.Role, error) {
	var role *model.Role
	if err := db.StockDB().WithContext(ctx).Table("role").Where("id = ?", id).Take(&role).Error; err != nil {
		return nil, err
	}
	return role, nil
}
