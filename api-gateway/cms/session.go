package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"stock/api-gateway/db"
	"stock/api-gateway/model"
	"stock/api-gateway/serr"
	"stock/api-gateway/util"
	"stock/common/log"
	"strings"
	"time"
)

// UserSession 用户session
type UserSession struct {
	SessionID  string      `json:"session_id"`
	CreateTime string      `json:"create_time"`
	Role       *model.Role `json:"role"`
}

func Session(ctx context.Context, role *model.Role) (string, error) {
	us := getUserSession(ctx, role.UserName)
	if us != nil {
		return us.SessionID, nil
	}
	sessionID := genSessionID(role)
	us = &UserSession{
		SessionID:  sessionID,
		CreateTime: time.Now().Format("2006-01-02 15:04:05"),
		Role:       role,
	}
	data, err := json.Marshal(us)
	if err != nil {
		return "", err
	}
	if err := db.RedisClient().Set(ctx, getUserSessionKey(role.UserName), string(data), maxExpireTTL*time.Second).Err(); err != nil {
		return "", err
	}
	return sessionID, nil
}

func getUserSessionKey(username string) string {
	return fmt.Sprintf(userSessionLayoutKey, username)
}

// genSessionID 生成sessionID
func genSessionID(role *model.Role) string {
	// 使用 username + 时间戳生成 session id
	randomStr := util.MD5(fmt.Sprintf("%s_%s_%d", role.UserName, randomStrSalt, time.Now().Unix()))
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s_%s", role.UserName, randomStr)))
}

// getUserSession 查询用户session
func getUserSession(ctx context.Context, userName string) *UserSession {
	session := db.RedisClient().Get(ctx, getUserSessionKey(userName)).Val()
	if session == "" {
		return nil
	}
	var us UserSession
	if err := json.Unmarshal([]byte(session), &us); err != nil {
		return nil
	}
	return &us
}

func GetSession(ctx context.Context, token string) (*model.Role, error) {
	result, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		log.Errorf("DecodeString err:%+v", err)
		return nil, err
	}
	list := strings.Split(string(result), "_")
	if len(list) != 2 {
		return nil, errors.New("invalid token")
	}
	us := getUserSession(ctx, list[0])
	if us == nil {
		return nil, serr.ErrBusiness("用户session未找到")
	}
	// 检查时间
	createTime, err := time.ParseInLocation("2006-01-02 15:04:05", us.CreateTime, time.Local)
	if err != nil {
		return nil, err
	}
	// 如果过期,则直接返回错误
	if time.Since(createTime).Seconds() > maxExpireTTL {
		return nil, serr.ErrBusiness("超长时间未登录,请重新登录")
	}
	return us.Role, nil
}
