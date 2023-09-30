package service

import (
	"os"
	"testing"

	"stock/api-gateway/db"
	"stock/common/env"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	//	e, err := env.LoadEnv("../conf/test.json")
	e, err := env.LoadEnv("../conf/test.json")
	if err != nil {
		panic(err)
	}
	env.SetGlobalEnv(e)
	db.InitRedisClient()
}
