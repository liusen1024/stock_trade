package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/common/log"

	"gorm.io/gorm"
)

// SmsDao 短信
type SmsDao struct{}

var _smsDao = &SmsDao{}

// SmsDaoInstance 提供一个可用的对象
func SmsDaoInstance() *SmsDao {
	return _smsDao
}

func save(ctx context.Context, phone, code string) error {
	sms := &model.Sms{
		Phone:  phone,
		Code:   code,
		Status: 1, // '状态:1生效 2失效'
	}
	return db.StockDB().WithContext(ctx).Create(sms).Error
}

// VerifySms 验证 手机号码+验证码未被验证过 true:验证通过 false:验证失败
func (s *SmsDao) VerifySms(ctx context.Context, phone string, code string) (bool, error) {
	var sms *model.Sms
	if err := db.StockDB().WithContext(ctx).
		Table("sms").Where("phone = ? and code = ? and status = 1 and DATE(send_time)=CURRENT_DATE and (time_to_sec(CURRENT_TIME)-time_to_sec(send_time)) < 600", phone, code).
		Take(&sms).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// InvalidCode 设置验证码失效
func (s *SmsDao) InvalidCode(ctx context.Context, phone string) error {
	err := db.StockDB().WithContext(ctx).Table("sms").Where("phone = ? ", phone).Update("status", model.SmsStatusInValid).Error
	if err != nil {
		log.Errorf("更新验证码状态失败 err:%+v", err)
		return err
	}
	return nil
}

// GetLatestSms 获取最近的一条短信
func (s *SmsDao) GetLatestSms(ctx context.Context, phone string) (*model.Sms, error) {
	var sms *model.Sms
	if err := db.StockDB().WithContext(ctx).Table("sms").Where("phone = ?", phone).Order("send_time desc").Limit(1).Take(&sms).Error; err != nil {
		return nil, err
	}
	return sms, nil
}

// Save 保存发送短信
func (s *SmsDao) Save(ctx context.Context, sms *model.Sms) error {
	return db.StockDB().WithContext(ctx).Create(sms).Error
}
