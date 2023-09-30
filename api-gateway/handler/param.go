package handler

import (
	"context"
	"fmt"
	"stock/api-gateway/dao"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/service"
	"stock/api-gateway/util"
	"stock/common/log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// String 获取string格式参数
func String(c *gin.Context, key string) (string, error) {
	value := c.Request.Form.Get(key)
	if len(value) == 0 {
		return "", serr.New(serr.ErrCodeInvalidParam, fmt.Sprintf("param 【%s】 not set", key))
	}
	return value, nil
}

// StringWithException 获取string格式参数
func StringWithException(c *gin.Context, key, errMsg string) (string, error) {
	value := c.Request.Form.Get(key)
	if len(value) == 0 {
		return "", serr.New(serr.ErrCodeInvalidParam, fmt.Sprintf("%s", errMsg))
	}
	return value, nil
}

// StringWithDefault 获取string格式参数，若未赋值，则返回默认值
func StringWithDefault(c *gin.Context, key, defaultValue string) string {
	value, err := String(c, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// Int 获取int格式参数
func Int(c *gin.Context, key string) (int, error) {
	value := c.Request.Form.Get(key)
	if len(value) == 0 {
		return 0, serr.Errorf(serr.ErrCodeInvalidParam, "param %s not set", key)
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return 0, serr.Errorf(serr.ErrCodeInvalidParam, "param %s is not int type", key)
	}
	return v, nil
}

// Bool 获取bool格式参数
func Bool(c *gin.Context, key string) (bool, error) {
	value := c.Request.Form.Get(key)
	if len(value) == 0 {
		return false, serr.Errorf(serr.ErrCodeInvalidParam, "param %s not set", key)
	}
	v, err := strconv.ParseBool(value)
	if err != nil {
		return false, serr.Errorf(serr.ErrCodeInvalidParam, "param %s is not bool type", key)
	}
	return v, nil
}

// BoolWithDefault 获取bool格式参数，未赋值则返回默认
func BoolWithDefault(c *gin.Context, key string, defaultValue bool) bool {
	value, err := Bool(c, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// Int64 获取int64格式参数
func Int64(c *gin.Context, key string) (int64, error) {
	value := c.Request.Form.Get(key)
	if len(value) == 0 {
		return 0, serr.Errorf(serr.ErrCodeInvalidParam, "param [%s] not set", key)
	}
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, serr.Errorf(serr.ErrCodeInvalidParam, "param [%s] is not Int type", key)
	}
	return v, nil
}

// Float64 获取float64格式参数
func Float64(c *gin.Context, key string) (float64, error) {
	value := c.Request.Form.Get(key)
	if len(value) == 0 {
		return 0, serr.Errorf(serr.ErrCodeInvalidParam, "param 【%s】 not set", key)
	}
	v, err := strconv.ParseFloat(value, 10)
	if err != nil {
		return 0, serr.Errorf(serr.ErrCodeInvalidParam, "param %s is not Int type", key)
	}
	return v, nil
}

// Int64WithDefault 获取int64格式参数，未赋值返回默认值
func Int64WithDefault(c *gin.Context, key string, defaultValue int64) int64 {
	value, err := Int64(c, key)
	if err != nil {
		return defaultValue
	}
	return value
}

// UserID 获取userID
func UserID(c *gin.Context) (int64, error) {
	uid, err := Int64(c, "uid")
	if err != nil {
		return 0, serr.New(serr.ErrCodeNoLogin, "请登录")
	}
	// 检查用户是否允许交易
	user, err := dao.UserDaoInstance().GetUserByUID(util.RPCContext(c), uid)
	if err != nil {
		return 0, serr.Errorf(serr.ErrCodeNoLogin, "请登录")
	}
	if user.Status != model.UserStatusActive {
		return 0, serr.New(serr.ErrCodeBusinessFail, "用户已被冻结")
	}
	// 在线
	if _, err := db.Set(context.Background(), fmt.Sprintf("user_online_%d", uid), "1", 30*time.Second).Result(); err != nil {
		log.Errorf("UserID online:%+v", err)
	}
	return uid, nil
}

// PositionID 获取委托编号
func PositionID(c *gin.Context) (int64, error) {
	v, err := Int64(c, "position_id")
	if err != nil {
		return 0, err
	}
	return v, nil
}

// Price 获取价格格式参数
func Price(c *gin.Context) (float64, error) {
	value := c.Request.Form.Get("price")
	if len(value) == 0 {
		return 0, nil
	}
	v, err := strconv.ParseFloat(value, 10)
	if err != nil {
		return 0, serr.Errorf(serr.ErrCodeBusinessFail, "价格错误")
	}
	return v, nil
}

// Amount 数量
func Amount(c *gin.Context) (int64, error) {
	value := c.Request.Form.Get("amount")
	if len(value) == 0 {
		return 0, serr.Errorf(serr.ErrCodeBusinessFail, "委托失败:交易股数错误")
	}
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, serr.Errorf(serr.ErrCodeBusinessFail, "委托失败:交易股数错误")
	}
	return v, nil
}

// ContractID 获取合约ID
func ContractID(c *gin.Context) (int64, error) {
	value, err := Int64(c, "contract_id")
	if err != nil {
		return 0, serr.ErrBusiness("合约不存在")
	}
	_, err = dao.ContractDaoInstance().GetContractByID(util.RPCContext(c), value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// EntrustID 获取委托id
func EntrustID(c *gin.Context) (int64, error) {
	value, err := Int64(c, "entrust_id")
	if err != nil {
		return 0, err
	}
	return value, nil
}

// StockCode 获取股票代码
func StockCode(c *gin.Context) (string, error) {
	value, err := String(c, "code")
	if err != nil {
		return "", serr.New(serr.ErrCodeBusinessFail, "请输入股票代码")
	}
	_, err = service.StockDataServiceInstance().GetStockDataByCode(util.RPCContext(c), value)
	if err != nil {
		return "", err
	}
	return value, nil
}

// EntrustProp 委托类型:1限价 2市价
func EntrustProp(c *gin.Context) (int64, error) {
	value, err := Int64(c, "type")
	if err != nil {
		return 0, serr.New(serr.ErrCodeBusinessFail, "委托类型错误")
	}
	switch value {
	case 1:
		return model.EntrustPropTypeLimitPrice, nil
	case 0:
		return model.EntrustPropTypeMarketPrice, nil
	}
	return value, nil
}

// SortBy 获取string格式参数，若未赋值，则返回默认值
func SortBy(c *gin.Context) (string, error) {
	value, err := String(c, "sort_by")
	if err != nil {
		return "", nil
	}
	if len(value) > 0 && (value != "price" && value != "chg_percent") {
		return "", serr.ErrBusiness("只能选择price|chg_percent排序")
	}
	return value, nil
}

// OrderBy 获取string格式参数，若未赋值，则返回默认值
func OrderBy(c *gin.Context) (string, error) {
	value, err := String(c, "order_by")
	if err != nil {
		return "", nil
	}
	if len(value) > 0 && (value != "asc" && value != "desc") {
		return "", serr.ErrBusiness("asc|desc排序")
	}
	return value, nil
}
