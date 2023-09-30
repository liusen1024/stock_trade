package main

import (
	"context"
	"flag"
	cms "stock/api-gateway/cms"
	"stock/api-gateway/db"
	"stock/api-gateway/handler"
	hq "stock/api-gateway/hq"
	"stock/api-gateway/mw"
	"stock/api-gateway/service"
	task "stock/api-gateway/task"
	"stock/common/env"
	"stock/common/log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 启动参数
	var conf string
	flag.StringVar(&conf, "conf", "conf/conf.json", "指定启动配置文件")
	flag.Parse()
	// load env config
	env.LoadGlobalEnv(conf)
	ctx := context.Background()
	// init redis client
	db.Init(ctx)
	db.InitRedisClient()

	engine := gin.Default()
	engine.Use(mw.Recovery)
	engine.Use(mw.ParseFormMiddleware)
	engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// 注册其他的API
	handler.Register(engine)
	hq.Register(engine)
	cms.Register(engine)

	// 预加载实例
	service.Init()

	task.TaskServiceInstance()

	err := engine.Run()
	log.Errorf("gin engine failed: %v", err)
}
