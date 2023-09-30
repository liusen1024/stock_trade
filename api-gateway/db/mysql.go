package db

import (
	"context"
	"errors"
	"stock/common/env"
	"stock/common/log"
	"time"

	mysql2 "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Init 数据库初始化连接
func Init(ctx context.Context) {
	dsn, ok := env.GlobalEnv().Get("STOCKDB")
	if !ok {
		panic("no TRADEDB env set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic(ctx, "connect to db err", err)
	}

	// 连接池设置
	sqlDB, err := db.DB()
	if err != nil {
		log.Panic(ctx, "connect to db err", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Infof("connect to db trade db success")
	stockDB = db
}

var stockDB *gorm.DB

// StockDB return cond database
func StockDB() *gorm.DB {
	return stockDB
}

// IsErrDuplicateKey test error
func IsErrDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql2.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
