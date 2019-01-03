package server

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"github.com/soooooooon/rock/models"
	"github.com/soooooooon/rock/session"
	melody "gopkg.in/olahol/melody.v1"
)

func handleWebsocketRequests(ctx context.Context) gin.HandlerFunc {
	m := melody.New()
	m.Config.PingPeriod = 30 * time.Second
	m.Config.PongWait = 3 * time.Minute

	go func() {
		if err := pullBetRecords(ctx, m); err != nil {
			log.Errorf("pull records failed: %s", err)
		}
	}()

	return func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	}
}

func pullBetRecords(ctx context.Context, m *melody.Melody) error {
	lastRecord := &models.Record{}
	db := session.MysqlRead(ctx).Last(lastRecord)
	if db.Error != nil && !db.RecordNotFound() {
		return db.Error
	}

	req := &models.RecordRequest{
		FromId: lastRecord.ID,
	}

	const (
		limit = 10
		wait  = time.Second
	)

	for {
		records, err := models.QueryRecords(ctx, req, limit)
		if err != nil {
			log.Errorf("query records failed: %s", err)
			time.Sleep(wait)
			continue
		}

		updatedArenas := map[uint]*models.Record{}
		for _, r := range records {
			if r.Err == 0 {
				updatedArenas[r.ArenaId] = r
			}

			req.FromId = r.ID
		}

		for id, record := range updatedArenas {
			arena, err := models.ArenaWithId(ctx, id)
			if err != nil {
				return err
			}

			body := map[string]interface{}{
				"id": models.HashIdEncode(id),
				"a":  arena.Amount,
				"b":  arena.Balance,
				"r": map[string]interface{}{
					"id": record.ID,
					"a":  record.Amount,
					"r":  record.Reward,
					"g":  record.Gestures,
					"G":  record.DefendGestures,
				},
			}

			msg, _ := jsoniter.Marshal(body)
			if err := m.Broadcast(msg); err != nil {
				log.Errorf("brodcast msg failed: %s", err)
			}
		}

		if len(records) < limit {
			time.Sleep(wait)
		}
	}
}
