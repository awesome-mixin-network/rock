package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/soooooooon/rock/session"
)

type Property struct {
	Key   string `gorm:"type:varchar(128);PRIMARY_KEY"`
	Value string `gorm:"type:varchar(256);"`
}

func ReadProperty(ctx context.Context, key string) (string, error) {
	p := Property{Key: key}
	db := session.MysqlRead(ctx).Where(p).First(&p)
	if db.RecordNotFound() {
		return "", nil
	}

	return p.Value, db.Error
}

func WriteProperty(ctx context.Context, key, value string) error {
	p := Property{Key: key}
	db := session.MysqlWrite(ctx)
	return db.Where(p).Assign(Property{Value: value}).FirstOrCreate(&p).Error
}

func ReadPropertyAsTime(ctx context.Context, key string) (time.Time, error) {
	value, err := ReadProperty(ctx, key)
	if err != nil {
		return time.Time{}, err
	}

	t, _ := time.Parse(time.RFC3339Nano, value)
	return t, nil
}

func WritePropertyAsTime(ctx context.Context, key string, value time.Time) error {
	return WriteProperty(ctx, key, value.UTC().Format(time.RFC3339Nano))
}

func ReadPropertyAsUint(ctx context.Context, key string) (uint, error) {
	value, err := ReadProperty(ctx, key)
	if err != nil {
		return 0, err
	}

	u64, _ := strconv.ParseUint(value, 10, 32)
	return uint(u64), nil
}

func WritePropertyAsUint(ctx context.Context, key string, value uint) error {
	return WriteProperty(ctx, key, fmt.Sprintf("%d", value))
}
