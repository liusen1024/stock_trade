package dao

import (
	"context"
	"errors"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"

	"gorm.io/gorm/clause"

	"gorm.io/gorm"
)

// UserDao 用户
type UserDao struct{}

var _userDao = &UserDao{}

// UserDaoInstance 提供一个可用的对象
func UserDaoInstance() *UserDao {
	return _userDao
}

// GetUserByUserName 根据用户名查找用户
func (s *UserDao) GetUserByUserName(ctx context.Context, userName string) (*model.User, error) {
	var user *model.User
	err := db.StockDB().WithContext(ctx).Table("users").Where(
		"user_name = ?", userName).Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 空记录
			return nil, gorm.ErrRecordNotFound
		}
		log.Errorf("系统错误:查询user失败")
		return nil, err
	}
	return user, nil
}

// GetUserByUID 根据UID查询用户
func (s *UserDao) GetUserByUID(ctx context.Context, uid int64) (*model.User, error) {
	var user *model.User
	sql := "select * from users where id = ?"
	err := db.StockDB().WithContext(ctx).Raw(sql, uid).Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 空记录
			return nil, serr.New(serr.ErrCodeBusinessFail, "用户不存在")
		}
		log.Errorf("查询用户失败:uid[%+v],err:%+v", uid, err)
		return nil, serr.New(serr.ErrCodeBusinessFail, "查询用户失败")
	}
	return user, nil
}

// SetCurrentContract 设置用户当前合约
func (s *UserDao) SetCurrentContract(ctx context.Context, uid, contractID int64) error {
	sql := "update users set current_contract_id = ? where id = ?"
	err := db.StockDB().WithContext(ctx).Exec(sql, contractID, uid).Error
	if err != nil {
		log.Errorf("更新用户当前合约失败:uid[%v] current_contract_id[%v]", uid, contractID)
		return serr.New(serr.ErrCodeBusinessFail, "更新用户当前合约失败")
	}
	return nil
}

// CreateUser 注册用户
func (s *UserDao) CreateUser(ctx context.Context, user *model.User) error {
	if err := db.StockDB().WithContext(ctx).Table("users").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&user).Error; err != nil {
		return err
	}
	return nil
}

// UpdateUserWithTx 更新用户
func (s *UserDao) UpdateUserWithTx(tx *gorm.DB, user *model.User) error {
	if err := tx.Table("users").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&user).Error; err != nil {
		log.Errorf("UpdateUserWithTx err:%+v", err)
		return err
	}
	return nil
}

func (s *UserDao) ModifyUserPassword(ctx context.Context, userName, newPassword string) error {
	err := db.StockDB().WithContext(ctx).Table("users").Where("user_name = ? ", userName).Update("password", newPassword).Error
	if err != nil {
		log.Errorf("修改密码失败 err:%+v", err)
		return err
	}
	return nil
}

// CheckUserExist 检查用户是否存在,存在返回true,不存在返回false
func (s *UserDao) CheckUserExist(ctx context.Context, uid int64) error {
	var user *model.User
	sql := `select * from users where id = ?`
	err := db.StockDB().WithContext(ctx).Raw(sql, uid).Take(&user).Error
	if err != nil || user.ID == 0 {
		return serr.Errorf(serr.ErrCodeNoLogin, "请登录")
	}
	return nil
}

// UpdateCurrentContractID 更新用户当前合约id
func (s *UserDao) UpdateCurrentContractID(ctx context.Context, uid, contractID int64) error {
	sql := "update users set current_contract_id = ? where id = ?"
	if err := db.StockDB().WithContext(ctx).Exec(sql, contractID, uid).Error; err != nil {
		log.Errorf("更新用户当前合约ID失败:%+v", err)
		return serr.ErrBusiness("更新合约失败")
	}
	return nil
}

func (s *UserDao) GetUserByRoleIDs(ctx context.Context, roleIDs []int64) ([]*model.User, error) {
	var list []*model.User
	if err := db.StockDB().WithContext(ctx).Table("users").Where("role_id in (?)", roleIDs).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *UserDao) GetUsers(ctx context.Context) ([]*model.User, error) {
	var list []*model.User
	if err := db.StockDB().WithContext(ctx).Table("users").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
