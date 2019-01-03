package session

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

type contextKey int

const (
	_ contextKey = iota
	mysqlReadKey
	mysqlWriteKey
	redisClientKey
)

func WithMysqlRead(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, mysqlReadKey, db)
}

func WithMysqlWrite(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, mysqlWriteKey, db)
}

func WithRedisClient(ctx context.Context, c *redis.Client) context.Context {
	return context.WithValue(ctx, redisClientKey, c)
}

func MysqlRead(ctx context.Context) *gorm.DB {
	return ctx.Value(mysqlReadKey).(*gorm.DB)
}

func MysqlWrite(ctx context.Context) *gorm.DB {
	return ctx.Value(mysqlWriteKey).(*gorm.DB)
}

func WithMysqlBegin(ctx context.Context) context.Context {
	db := MysqlWrite(ctx).Begin()
	ctx = WithMysqlWrite(ctx, db)
	ctx = WithMysqlRead(ctx, db)
	return ctx
}

func MysqlRollback(ctx context.Context) *gorm.DB {
	return MysqlWrite(ctx).Rollback()
}

func MysqlCommit(ctx context.Context) *gorm.DB {
	return MysqlWrite(ctx).Commit()
}

func Redis(ctx context.Context) *redis.Client {
	return ctx.Value(redisClientKey).(*redis.Client)
}
