package dao

import (
	"context"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/common/log"

	"gorm.io/gorm"
)

// MsgDao 消息
type MsgDao struct{}

var _msgDao = &MsgDao{}

// MsgDaoInstance 提供一个可用的对象
func MsgDaoInstance() *MsgDao {
	return _msgDao
}

// Create 创建msg
func (s *MsgDao) Create(ctx context.Context, msg *model.Msg) error {
	if err := db.StockDB().WithContext(ctx).Table("msg").Create(&msg).Error; err != nil {
		log.Errorf("创建msg失败:%+v", err)
		return err
	}
	return nil
}

// CreateWithTx 创建msg
func (s *MsgDao) CreateWithTx(tx *gorm.DB, msg *model.Msg) error {
	if err := tx.Table("msg").Create(&msg).Error; err != nil {
		log.Errorf("创建msg失败:%+v", err)
		return err
	}
	return nil
}

// GetByUid 根据用户ID查询消息
func (s *MsgDao) GetByUid(ctx context.Context, uid int64) ([]*model.Msg, error) {
	var list []*model.Msg
	err := db.StockDB().WithContext(ctx).Table("msg").Where("uid = ?", uid).Find(&list).Error
	if err != nil {
		return nil, serr.New(serr.ErrCodeContractNoFound, "查询消息失败")
	}
	return list, nil
}
