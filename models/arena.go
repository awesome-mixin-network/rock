package models

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/soooooooon/rock/session"
	"github.com/vmihailenco/msgpack"
)

type Arena struct {
	ID         uint       `gorm:"PRIMARY_ID" json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"-"`
	DeletedAt  *time.Time `sql:"INDEX" json:"-"`
	ExpiredAt  time.Time  `sql:"INDEX" json:"expired_at"`
	ArchivedAt int64      `sql:"INDEX" json:"-"`

	UserId     string `sql:"type:varchar(36);INDEX" json:"user_id"`
	AssetId    string `sql:"type:varchar(36);INDEX" json:"asset_id"`
	SnapshotId string `sql:"type:varchar(36);" json:"-"`
	TraceId    string `sql:"type:varchar(36);UNIQUE_INDEX" json:"-"`
	Memo       string `sql:"type:varchar(140);" json:"-"`

	Amount     string `sql:"type:varchar(16);" json:"amount"`
	Balance    string `sql:"type:varchar(16);" json:"balance"`
	MaxBet     string `sql:"type:varchar(16);" json:"max_bet"`
	MinGesture int    `json:"min_gesture"`
	MaxGesture int    `json:"max_gesture"`

	Rank int64 `sql:"INDEX" json:"-"`
}

func ArenaWithId(ctx context.Context, id uint) (*Arena, error) {
	a := Arena{ID: id}
	db := session.MysqlRead(ctx).Where(a).First(&a)
	return &a, db.Error
}

func arenaCacheKey(traceId string) string {
	return "rock_arena_" + traceId
}

func (a *Arena) Cache(ctx context.Context) error {
	data, err := msgpack.Marshal(a)
	if err != nil {
		return err
	}

	key := arenaCacheKey(a.TraceId)
	_, err = session.Redis(ctx).Set(key, data, time.Hour*12).Result()
	return err
}

func ArenaFromCache(ctx context.Context, traceId string) (*Arena, error) {
	key := arenaCacheKey(traceId)
	data, err := session.Redis(ctx).Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	arena := Arena{}
	err = msgpack.Unmarshal(data, &arena)
	return &arena, err
}

type ArenaRequest struct {
	UserId  string
	AssetId string

	OnlyUnexpired bool

	FromId   uint
	ExceptId uint
}

func QueryArenas(ctx context.Context, req *ArenaRequest, limit int) ([]*Arena, error) {
	db := session.MysqlRead(ctx).Order("id DESC")

	if req != nil {
		if len(req.UserId) > 0 {
			db = db.Where("user_id = ?", req.UserId)
		}

		if len(req.AssetId) > 0 {
			db = db.Where("asset_id = ?", req.AssetId)
		}

		if req.ExceptId > 0 {
			db = db.Where("id <> ?", req.ExceptId)
		}

		if req.FromId > 0 {
			db = db.Where("id < ?", req.FromId)
		}

		if req.OnlyUnexpired {
			now := time.Now()
			db = db.Where("expired_at > ?", now)
		}
	}

	arenas := []*Arena{}
	err := db.Limit(limit).Find(&arenas).Error

	return arenas, err
}

func (a *Arena) UpdateBalance(ctx context.Context, amount decimal.Decimal) error {
	balance, _ := decimal.NewFromString(a.Balance)
	newBalance := balance.Add(amount).String()
	db := session.MysqlWrite(ctx).Model(a).Where("balance = ?", balance.String())
	return db.Update("balance", newBalance).Error
}

func ListPendingExpiredArenas(ctx context.Context, limit int) ([]Arena, error) {
	arenas := []Arena{}
	err := session.MysqlRead(ctx).
		Where("archived_at = 0 AND expired_at < ?", time.Now()).
		Limit(limit).
		Find(&arenas).Error

	return arenas, err
}

func ListTopRankArenas(ctx context.Context, limit int) ([]*Arena, error) {
	arenas := []*Arena{}
	err := session.MysqlRead(ctx).
		Where("expired_at > ?", time.Now()).
		Order("rank DESC").
		Limit(limit).
		Find(&arenas).Error

	return arenas, err
}
