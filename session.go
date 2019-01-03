package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/soooooooon/rock/config"
	"github.com/soooooooon/rock/session"
)

func openMysqlDB(host string) *gorm.DB {
	path := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=True&charset=utf8mb4",
		config.MysqlUserName,
		config.MysqlPassword,
		"tcp",
		host,
		config.MysqlDatabaseName,
	)

	db, err := gorm.Open("mysql", path)
	if err != nil {
		panic(err)
	}

	db.DB().SetMaxIdleConns(10)
	return db
}

func openRedisConn() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       0,
	})

	if _, err := client.Ping().Result(); err != nil {
		panic(err)
	}

	return client
}

func withMysql(ctx context.Context) context.Context {
	ctx = session.WithMysqlRead(ctx, openMysqlDB(config.MysqlReadHost))
	ctx = session.WithMysqlWrite(ctx, openMysqlDB(config.MysqlWriteHost))

	if log.GetLevel() == log.DebugLevel {
		session.MysqlRead(ctx).LogMode(true)
		session.MysqlWrite(ctx).LogMode(true)
	}

	return ctx
}

func withRedis(ctx context.Context) context.Context {
	ctx = session.WithRedisClient(ctx, openRedisConn())
	return ctx
}

func withSession(ctx context.Context) context.Context {
	ctx = withMysql(ctx)
	ctx = withRedis(ctx)
	return ctx
}
