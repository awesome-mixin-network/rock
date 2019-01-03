package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/soooooooon/rock/models"
	"github.com/soooooooon/rock/session"
)

func UpdateArenaRanks(ctx context.Context) error {
	assets, err := ReadAssets(ctx)
	if err != nil {
		return err
	}

	prices := map[string]decimal.Decimal{}
	for _, a := range assets {
		p, _ := decimal.NewFromString(a.PriceUSD)
		prices[a.AssetId] = p
	}

	var (
		fromId = uint(0)
		ranks  = map[uint]int64{}
	)

	for {
		req := &models.ArenaRequest{
			OnlyUnexpired: true,
			FromId:        fromId,
		}

		arenas, err := models.QueryArenas(ctx, req, 10)
		if err != nil {
			return err
		}

		ago := time.Now().Add(-24 * time.Hour)
		for _, a := range arenas {
			rank := a.CreatedAt.Unix()

			// query records
			db := session.MysqlRead(ctx).Model(&models.Record{}).Where("arena_id = ?", a.ID)
			if ago.After(a.CreatedAt) {
				db = db.Where("created_at > ?", ago)
			}

			var count uint = 0
			if err := db.Count(&count).Error; err != nil {
				return err
			}
			rank += int64(count) * 10 * 60

			balance, _ := decimal.NewFromString(a.Balance)
			price, _ := prices[a.AssetId]

			if price.IsPositive() {
				rank += balance.Mul(price).Mul(decimal.New(10*60, 0)).IntPart()
			} else {
				rank += balance.Shift(-4).IntPart()
			}

			ranks[a.ID] = rank
			fromId = a.ID
		}

		if len(arenas) < limit {
			break
		}
	}

	if len(ranks) > 0 {
		ctx := session.WithMysqlBegin(ctx)
		for id, rank := range ranks {
			if err := session.MysqlWrite(ctx).Model(&models.Arena{ID: id}).Update("rank", rank).Error; err != nil {
				session.MysqlRollback(ctx)
				return err
			}
		}
		if err := session.MysqlCommit(ctx).Error; err != nil {
			return err
		}
	}

	return nil
}

func HandlePendingExpiredArenas(ctx context.Context) error {
	const limit = 10

	arenas, err := models.ListPendingExpiredArenas(ctx, limit)
	if err != nil {
		return err
	}

	cutoff := decimal.New(2, -2) // 2%
	for _, a := range arenas {
		if b, _ := decimal.NewFromString(a.Balance); b.IsPositive() {
			amount, _ := decimal.NewFromString(a.Amount)
			if earnings := b.Sub(amount); earnings.IsPositive() {
				b = b.Sub(earnings.Mul(cutoff))
			}

			traceId := traceIdFromSnapshotId(a.SnapshotId)
			change := b.Sub(amount)
			memo := fmt.Sprintf(
				"arena %s, balance %s, change %s",
				models.HashIdEncode(a.ID),
				b.Truncate(8).String(),
				change.Truncate(8).String(),
			)

			if _, err := models.CreatePayment(ctx, traceId, a.AssetId, b.String(), a.UserId, memo); err != nil {
				return err
			}
		}

		if err := session.MysqlWrite(ctx).Model(&a).Update("archived_at", time.Now().Unix()).Error; err != nil {
			return err
		}
	}

	if len(arenas) >= limit {
		return HandlePendingExpiredArenas(ctx)
	}

	return nil
}
