package models

import (
	"context"
	"fmt"
	"time"

	"github.com/soooooooon/rock/session"
	"github.com/vmihailenco/msgpack"
)

const (
	Invalid = iota
	Lose
	Draw
	Win
)

type RecordErr int

const (
	_ RecordErr = iota
	InvalidArenaId
	InvalidAssetId
	ArenaExpired
	BeyondMaxBet
	NumberOfGesturesOutOfRange
	PossibleRewardBeyondBalance
)

func (err RecordErr) String() string {
	switch err {
	case InvalidArenaId:
		return "invalid arena id"
	case InvalidAssetId:
		return "asset not match"
	case ArenaExpired:
		return "arena closed"
	case BeyondMaxBet:
		return "beyond max bet limit"
	case NumberOfGesturesOutOfRange:
		return "number of gestures out of range"
	case PossibleRewardBeyondBalance:
		return "arena's balance is not enough to pay possible reawrd"
	}

	return ""
}

type Record struct {
	ID             uint      `sql:"PRIMARY_KEY" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	SnapshotId     string    `sql:"type:varchar(36);" json:"snapshot_id"`
	TraceId        string    `sql:"type:varchar(36);UNIQUE_INDEX" json:"-"` // trace id
	AssetId        string    `sql:"type:varchar(36);INDEX" json:"asset_id"`
	ArenaId        uint      `sql:"INDEX" json:"-"`
	UserId         string    `sql:"type:varchar(36);INDEX" json:"user_id"`
	Amount         string    `sql:"type:varchar(16);" json:"amount"`
	Memo           string    `sql:"type:varchar(140);" json:"-"`
	Result         int       `json:"result"`
	Err            RecordErr `sql:"INDEX" json:"err"`
	Gestures       string    `sql:"type:varchar(16);" json:"gestures"`
	DefendGestures string    `sql:"type:varchar(16);" json:"defend_gesture"`
	Reward         string    `sql:"type:varchar(16);" json:"reward"`
}

func RecordWithTraceId(ctx context.Context, traceId string) (*Record, error) {
	r := &Record{}
	db := session.MysqlRead(ctx).Where("trace_id = ?", traceId).Last(r)
	if db.RecordNotFound() {
		return nil, nil
	}

	if db.Error != nil {
		return nil, db.Error
	}

	return r, nil
}

func recordCacheKey(traceId string) string {
	return "rock_records_" + traceId
}

func (r *Record) Cache(ctx context.Context) error {
	data, err := msgpack.Marshal(r)
	if err != nil {
		return err
	}

	key := recordCacheKey(r.TraceId)
	_, err = session.Redis(ctx).Set(key, data, time.Hour*12).Result()
	return err
}

func RecordFromCache(ctx context.Context, traceId string) (*Record, error) {
	key := recordCacheKey(traceId)
	data, err := session.Redis(ctx).Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	record := Record{}
	err = msgpack.Unmarshal(data, &record)
	return &record, err
}

type RecordRequest struct {
	UserId     string
	ArenaId    uint
	AssetId    string
	FromId     uint
	Desc       bool
	IncludeErr bool
}

func QueryRecords(ctx context.Context, req *RecordRequest, limit int) ([]*Record, error) {
	db := session.MysqlRead(ctx)

	if req != nil {
		if len(req.UserId) > 0 {
			db = db.Where("user_id = ?", req.UserId)
		}

		if req.ArenaId > 0 {
			db = db.Where("arena_id = ?", req.ArenaId)
		}

		if len(req.AssetId) > 0 {
			db = db.Where("asset_id = ?", req.AssetId)
		}

		op := ">"
		if req.Desc {
			op = "<"
			db = db.Order("id DESC")
		}

		if req.FromId > 0 {
			db = db.Where(fmt.Sprintf("id %s %d", op, req.FromId))
		}

		if !req.IncludeErr {
			db = db.Where("err = 0")
		}
	}

	records := []*Record{}
	err := db.Limit(limit).Find(&records).Error

	return records, err
}
