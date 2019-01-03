package models

import (
	"context"

	"github.com/soooooooon/rock/session"
	hashids "github.com/speps/go-hashids"
)

var hashid *hashids.HashID = nil

func init() {
	hd := hashids.NewData()
	hd.Salt = "CGKw4duQirvKMUbTcZCZwywzjrzp"
	hd.MinLength = 4
	hashid, _ = hashids.NewWithData(hd)
}

func HashIdEncode(id uint) string {
	e, _ := hashid.EncodeInt64([]int64{int64(id)})
	return e
}

func HashIdDecode(id string) (uint, error) {
	d, err := hashid.DecodeInt64WithError(id)
	if err != nil {
		return 0, err
	}

	return uint(d[0]), nil
}

func SetDb(ctx context.Context) error {
	db := session.MysqlWrite(ctx).
		Set("gorm:table_options", "CHARSET=utf8mb4").
		AutoMigrate(
			&Arena{},
			&Record{},
			&Property{},
			&Payment{},
			&User{},
		)

	return db.Error
}
