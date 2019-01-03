package engine

import (
	"context"

	bot "github.com/MixinNetwork/bot-api-go-client"
	number "github.com/MixinNetwork/go-number"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/soooooooon/rock/config"
	"github.com/soooooooon/rock/models"
	"github.com/soooooooon/rock/session"
)

func traceIdFromSnapshotId(id string) string {
	uid, _ := uuid.FromString(id)
	return uuid.NewV5(uid, "refund or reward").String()
}

func handleUnpaidPayments(ctx context.Context) error {
	const limit = 10

	payments, err := models.UnpaidPayments(ctx, limit)
	if err != nil {
		return err
	}

	for _, p := range payments {
		input := &bot.TransferInput{
			AssetId:     p.AssetId,
			RecipientId: p.UserId,
			Amount:      number.FromString(p.Amount),
			TraceId:     p.TraceId,
			Memo:        p.Memo,
		}
		if err := bot.CreateTransfer(
			ctx,
			input,
			config.MixinClientId,
			config.MixinSessionId,
			config.MixinPrivateKey,
			config.MixinPin,
			config.MixinPinToken,
		); err != nil {
			log.Errorf("handle payment %d failed: %s", p.ID, err)
			return err
		}

		if db := session.MysqlWrite(ctx).Delete(&p); db.Error != nil && !db.RecordNotFound() {
			return db.Error
		}
	}

	if len(payments) >= limit {
		return handleUnpaidPayments(ctx)
	}

	return nil
}
