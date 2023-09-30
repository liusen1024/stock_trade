package service

import (
	"context"
	"gorm.io/gorm"
	"stock/api-gateway/dao"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"
	"sync"
	"time"
)

// UserService 内容服务
type UserService struct {
}

var (
	userService *UserService
	userOnce    sync.Once
)

// UserServiceInstance UserServiceInstance用户服务
func UserServiceInstance() *UserService {
	userOnce.Do(func() {
		userService = &UserService{}
	})
	return userService
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, userName, password string) (*model.User, error) {
	user, err := dao.UserDaoInstance().GetUserByUserName(ctx, userName)
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, serr.ErrBusiness("账号不存在")
	}
	if user.Password != password {
		return nil, serr.ErrBusiness("密码错误")
	}
	if user.Status != model.UserStatusActive {
		return nil, serr.ErrBusiness("账户被冻结")
	}

	return user, nil
}

// RegisterUser 注册用户
func (s *UserService) RegisterUser(ctx context.Context, userName, password, code, RegisterCode string) error {
	// 1.检查用户是否已经创建
	user, err := dao.UserDaoInstance().GetUserByUserName(ctx, userName)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Errorf("err:%+v", err)
		return serr.ErrBusiness("注册失败")
	}
	if user != nil {
		return serr.ErrBusiness("用户已存在")
	}

	// 2.检查验证码是否正确
	sms, err := dao.SmsDaoInstance().VerifySms(ctx, userName, code)
	if err != nil || sms == false {
		return serr.ErrBusiness("验证码错误")
	}

	// 3.检查是否需要推荐码(代理商) 需要场景:检查代理商是否存在 不存在则抛异常
	sys, err := dao.SysDaoInstance().GetSysParam(ctx)
	if err != nil {
		log.Errorf("err:%+v", err)
		return serr.ErrBusiness("注册失败")
	}

	var roleID int64
	if sys.RegistCode {
		if len(RegisterCode) == 0 {
			return serr.ErrBusiness("请填写推荐码")
		}
		role, err := dao.RoleDaoInstance().GetRoleByUserName(ctx, RegisterCode)
		if err != nil {
			return serr.ErrBusiness("请填写正确的推荐码")
		}
		roleID = role.ID
	}

	// 4.注册
	newUser := &model.User{
		UserName: userName,
		Password: password,
		Status:   model.UserStatusActive,
		RoleID:   roleID,
		CreateAt: time.Now(),
	}
	if err := dao.UserDaoInstance().CreateUser(ctx, newUser); err != nil {
		log.Errorf("err:%+v", err)
		return serr.ErrBusiness("注册失败")
	}

	// 5.验证码失效
	err = dao.SmsDaoInstance().InvalidCode(ctx, userName)
	if err != nil {
		log.Errorf("err:%+v", err)
	}

	// 新客户注册通知管理员
	if sys.RegisterNotice && len(sys.AdminPhone) > 0 {
		SmsServiceInstance().SendSms(ctx, "有新客户注册成功,请登录管理后台查看!", sys.AdminPhone)
	}
	return nil
}

// UpdatePassword 修改密码
func (s *UserService) UpdatePassword(ctx context.Context, userName, password, code string) error {
	//// 1.检查账户是否存在
	user, err := dao.UserDaoInstance().GetUserByUserName(ctx, userName)
	if err != nil && err == gorm.ErrRecordNotFound {
		log.Errorf("err:%+v", err)
		return serr.ErrBusiness("用户不存在")
	}
	if user.ID == 0 {
		return serr.ErrBusiness("用户不存在")
	}

	// 2.检查验证码是否正确
	sms, err := dao.SmsDaoInstance().VerifySms(ctx, userName, code)
	if err != nil || sms == false {
		return serr.ErrBusiness("验证码错误")
	}

	// 3.修改密码
	if err := dao.UserDaoInstance().ModifyUserPassword(ctx, userName, password); err != nil {
		return serr.ErrBusiness("修改密码失败")
	}

	// 4.设置验证码失效
	err = dao.SmsDaoInstance().InvalidCode(ctx, userName)
	if err != nil {
		log.Errorf("验证码设置失效失败:%+v", err)
	}
	return nil
}
