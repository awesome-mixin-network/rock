package models

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/soooooooon/rock/session"
)

type Payment struct {
	ID        uint       `gorm:"PRIMARY_ID" json:"id"`
	CreatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"INDEX" json:"-"`

	TraceId string `sql:"type:varchar(36);UNIQUE_INDEX" json:"-"`
	AssetId string `sql:"type:varchar(36)" json:"asset_id"`
	Amount  string `sql:"type:varchar(36)" json:"amount"`
	UserId  string `sql:"type:varchar(36);INDEX" json:"user_id"`
	Memo    string `sql:"type:varchar(140);" json:"memo"`
}

func UnpaidPayments(ctx context.Context, limit int) ([]Payment, error) {
	unpaid := []Payment{}
	err := session.MysqlRead(ctx).Order("id").Limit(limit).Find(&unpaid).Error
	return unpaid, err
}

func CreatePayment(ctx context.Context, traceId, assetId, amount, userId, memo string) (*Payment, error) {
	p := &Payment{
		TraceId: traceId,
		AssetId: assetId,
		Amount:  amount,
		UserId:  userId,
		Memo:    memo,
	}

	if _, err := uuid.FromString(p.UserId); err != nil {
		return nil, fmt.Errorf("invalid user id: %s", userId)
	}

	err := session.MysqlWrite(ctx).Where("trace_id = ?", traceId).FirstOrCreate(p).Error
	return p, err
}
