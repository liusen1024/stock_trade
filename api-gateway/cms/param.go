package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"stock/api-gateway/serr"
	"strconv"

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

func Username(c *gin.Context) string {
	v, ok := c.Get("__USERNAME")
	if ok {
		return v.(string)
	}
	data := map[string]interface{}{}
	ByteBody, _ := ioutil.ReadAll(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(ByteBody))
	if err := c.ShouldBind(&data); err != nil {
		return ""
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(ByteBody))
	v, ok = data["username"]
	if ok {
		return v.(string)
	}
	return ""
}

func IsAdmin(c *gin.Context) bool {
	return Username(c) == "admin"
}

func IsDownload(c *gin.Context) bool {
	ok, err := Bool(c, "download")
	if err != nil {
		return false
	}
	return ok
}

func SlicePage(c *gin.Context, count int) (int32, int32) {
	limit, err := Int64(c, "limit") // 固定的10条数据
	if err != nil {
		limit = 10
	}
	offset, err := Int64(c, "offset")
	if err != nil {
		offset = 1
	}

	start := offset - 1
	if start >= int64(count) {
		return 0, 0
	}
	end := start + limit
	if end >= int64(count) {
		end = int64(count)
	}
	return int32(start), int32(end)
}
