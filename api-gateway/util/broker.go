package util

import (
	"errors"
	"stock/common/env"
)

func BrokerHost() (string, error) {
	host, ok := env.GlobalEnv().Get("BROKER_HOST")
	if !ok {
		return "", errors.New("获取BROKER_HOST配置失败")
	}
	return host, nil
}
