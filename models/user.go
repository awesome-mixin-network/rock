package models

import (
	"context"
	"time"

	"github.com/soooooooon/rock/session"
	"github.com/vmihailenco/msgpack"
)

type User struct {
	ID      uint   `gorm:"PRIMARY_ID" json:"-"`
	MixinID string `gorm:"type:varchar(36);UNIQUE" json:"id"`
	Name    string `gorm:"type:varchar(36)" json:"name"`
	Avatar  string `gorm:"type:varchar(256)" json:"avatar"`

	RefreshedAt time.Time `gorm:"INDEX" json:"-"`
}

func ListOutdatedUsers(ctx context.Context, limit int) ([]*User, error) {
	date := time.Now().AddDate(0, 0, -7)
	users := []*User{}
	err := session.MysqlRead(ctx).
		Order("id DESC").
		Where("refreshed_at < ?", date).
		Limit(limit).
		Find(&users).Error

	return users, err
}

func (u *User) Cache(ctx context.Context) error {
	data, err := msgpack.Marshal(u)
	if err != nil {
		return err
	}

	key := "user_" + u.MixinID
	_, err = session.Redis(ctx).Set(key, data, time.Hour*24*3).Result()
	return err
}

func GetUserWithMixinID(ctx context.Context, mixinID string) (*User, error) {
	key := "user_" + mixinID
	user := &User{}
	if data, err := session.Redis(ctx).Get(key).Bytes(); err == nil {
		if err := msgpack.Unmarshal(data, &user); err == nil {
			return user, nil
		}
	}

	if err := session.MysqlRead(ctx).Where("mixin_id = ?", mixinID).First(&user).Error; err != nil {
		return nil, err
	}

	user.Cache(ctx)
	return user, nil
}

func FirstOrCreateUser(ctx context.Context, mixinID string) (*User, error) {
	user := User{MixinID: mixinID}
	err := session.MysqlWrite(ctx).Where(user).FirstOrCreate(&user).Error
	return &user, err
}
