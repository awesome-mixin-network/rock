package engine

import (
	"context"
	"fmt"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/soooooooon/rock/memo"
	"github.com/soooooooon/rock/models"
	"github.com/soooooooon/rock/session"
)

const (
	checkpointKey = "rock_bot_mixin_network"
	limit         = 500
	pullInterval  = 500 * time.Millisecond
	retryInterval = time.Second
)

type serviceFunc func(ctx context.Context) error

func runService(ctx context.Context, name string, f serviceFunc, interval int64) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			if interval > 1 && t.Unix()%interval != 0 {
				continue
			}

			if err := f(ctx); err != nil {
				log.Errorf("%s failed: %s", name, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func LaunchEngine(ctx context.Context, fromNow bool) error {
	// run services
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go runService(ctx, "update ranks", UpdateArenaRanks, 60)
	go runService(ctx, "handle pending expired arenas", HandlePendingExpiredArenas, 60)
	go runService(ctx, "handle unpaid payments", handleUnpaidPayments, 1)
	go runService(ctx, "update user profile", handleUpdateUsers, 1)

	checkpoint, err := models.ReadPropertyAsTime(ctx, checkpointKey)
	if err != nil {
		return err
	}

	if checkpoint.IsZero() || fromNow {
		checkpoint = time.Now().UTC()
	}

	snapshotFilter := map[string]bool{}
	assetFilter := map[string]bool{}

	for {
		// limit = 500
		snapshots, err := requestMixinNetwork(ctx, checkpoint, limit)
		if err != nil {
			log.Errorf("pull mixin network failed: %s", err)
			time.Sleep(retryInterval)
			continue
		}

		for _, s := range snapshots {
			checkpoint = s.CreatedAt

			if s.UserId == "" || s.TraceId == "" {
				continue
			}

			if _, ok := snapshotFilter[s.SnapshotId]; ok {
				continue
			}

			amount, _ := decimal.NewFromString(s.Amount)
			if !amount.IsPositive() {
				continue
			}

			if _, err := uuid.FromString(s.SnapshotId); err != nil {
				continue
			}

			if _, ok := assetFilter[s.Asset.AssetId]; !ok {
				if err := s.Asset.Cache(ctx); err != nil {
					log.Errorf("cache asset failed: %s", err)
				} else {
					assetFilter[s.Asset.AssetId] = true
				}
			}

			for {
				user, err := models.FirstOrCreateUser(ctx, s.OpponentId)
				if err != nil {
					log.Errorf("ensue user %s failed: %s", s.OpponentId, err)
					time.Sleep(retryInterval)
					continue
				}

				if err := processSnapshot(ctx, s, user); err != nil {
					log.Errorf("process snapshot %s failed: %s", s.SnapshotId, err)
					time.Sleep(retryInterval)
					continue
				}

				break
			}

			snapshotFilter[s.SnapshotId] = true
		}

		if err := models.WritePropertyAsTime(ctx, checkpointKey, checkpoint); err != nil {
			log.Errorf("update checkpoint failed: %s", err)
			time.Sleep(retryInterval)
			continue
		}

		if len(snapshots) < limit {
			time.Sleep(pullInterval)
		}
	}
}

func processSnapshot(ctx context.Context, s *Snapshot, user *models.User) error {
	if action, err := memo.NewArena(s.Data); err == nil {
		// create arena
		log.Infof("NEW CREATE ARENA ACTION: %#v", action)

		amount, _ := decimal.NewFromString(s.Amount)
		max, _ := decimal.NewFromString(action.M)
		if max.GreaterThan(amount) {
			max = amount
		}

		a := &models.Arena{
			ExpiredAt:  time.Now().Add(time.Duration(action.E) * time.Hour),
			UserId:     s.OpponentId,
			AssetId:    s.Asset.AssetId,
			SnapshotId: s.SnapshotId,
			TraceId:    s.TraceId,
			Amount:     amount.String(),
			Balance:    amount.String(),
			MaxBet:     max.String(),
			MinGesture: 1,
			MaxGesture: 4,
		}

		if err := session.MysqlWrite(ctx).Where("trace_id = ?", s.TraceId).FirstOrCreate(a).Error; err != nil {
			return err
		}

		return a.Cache(ctx)
	}

	if action, err := memo.NewBet(s.Data); err == nil {
		// new bet
		log.Infof("NEW BET ACTION: %#v", action)

		if r, err := models.RecordWithTraceId(ctx, s.TraceId); err != nil {
			return err
		} else if r != nil {
			return r.Cache(ctx)
		}

		r, err := recordFromSnapshot(ctx, s, action)
		if err != nil {
			return err
		}

		tsc := session.WithMysqlBegin(ctx)

		if r.Err != 0 {
			if err := refundSnapshot(tsc, s, r.Err.String()); err != nil {
				session.MysqlRollback(tsc)
				return err
			}
		} else {
			a, err := models.ArenaWithId(tsc, r.ArenaId)
			if err != nil {
				session.MysqlRollback(tsc)
				return err
			}

			amount, _ := decimal.NewFromString(r.Amount)

			if r.Result == models.Lose {
				if err := a.UpdateBalance(tsc, amount); err != nil {
					session.MysqlRollback(tsc)
					return err
				}
			} else if r.Result == models.Win {
				reward, _ := decimal.NewFromString(r.Reward)
				if err := a.UpdateBalance(tsc, amount.Sub(reward)); err != nil {
					session.MysqlRollback(tsc)
					return err
				}

				// fee 2%
				r.Reward = reward.Mul(decimal.New(98, -2)).String()
				if err := rewardRecord(tsc, r, "Congratulations！You did it."); err != nil {
					session.MysqlRollback(tsc)
					return err
				}
			} else {
				// 打平，退款
				if err := refundSnapshot(tsc, s, "Draw! Good luck next time."); err != nil {
					session.MysqlRollback(tsc)
					return err
				}
			}
		}

		if err := session.MysqlWrite(tsc).Create(r).Error; err != nil {
			session.MysqlRollback(tsc)
			return err
		}

		if err := session.MysqlCommit(tsc).Error; err != nil {
			return err
		}

		return r.Cache(ctx)
	}

	if action, err := memo.NewLogin(s.Data); err == nil {
		t := time.Unix(action.T, 0)
		if since := time.Since(t); since < 0 || since > time.Minute {
			return refundSnapshot(ctx, s, "expired login request")
		}

		l := models.Session{
			CreatedAt: time.Now(),
			UserId:    s.OpponentId,
			Token:     s.TraceId,
		}

		if err := l.Cache(ctx); err != nil {
			return err
		}

		return refundSnapshot(ctx, s, fmt.Sprintf("Welcome to ROCK! %s (%s)", l.UserId, l.Token))
	}

	switch strings.ToLower(s.Data) {
	case "deposit", "donate":
		// thx
		return nil
	default:
		return refundSnapshot(ctx, s, "unknown action")
	}
}

func recordFromSnapshot(ctx context.Context, s *Snapshot, action *memo.Bet) (*models.Record, error) {
	amount, _ := decimal.NewFromString(s.Amount)

	r := &models.Record{
		SnapshotId: s.SnapshotId,
		TraceId:    s.TraceId,
		AssetId:    s.Asset.AssetId,
		UserId:     s.OpponentId,
		Amount:     amount.String(),
		Memo:       s.Data,
	}

	id, err := models.HashIdDecode(action.A)
	if err != nil {
		r.Err = models.InvalidArenaId
		return r, nil
	}

	a, err := models.ArenaWithId(ctx, id)
	if err != nil {
		r.Err = models.InvalidArenaId
		return r, nil
	}

	r.ArenaId = a.ID

	if a.AssetId != s.Asset.AssetId {
		r.Err = models.InvalidAssetId
		return r, nil
	}

	if time.Now().After(a.ExpiredAt) {
		r.Err = models.ArenaExpired
		return r, nil
	}

	maxBet, _ := decimal.NewFromString(a.MaxBet)
	if amount.GreaterThan(maxBet) {
		r.Err = models.BeyondMaxBet
		return r, nil
	}

	gestures := models.NewGesturesFromString(action.G)
	numberOfGestures := len(gestures)
	if numberOfGestures < a.MinGesture || numberOfGestures > a.MaxGesture {
		r.Err = models.NumberOfGesturesOutOfRange
		return r, nil
	}

	pow := decimal.New(int64(2<<uint(numberOfGestures-1)), 0)
	maxReward := amount.Mul(pow)

	balance, _ := decimal.NewFromString(a.Balance)
	if balance.Add(amount).LessThan(maxReward) {
		r.Err = models.PossibleRewardBeyondBalance
		return r, nil
	}

	reward := amount
	snapshot, _ := uuid.FromString(s.SnapshotId)
	defendGestures := []int{}
	for idx, g := range gestures {
		d := models.NewGestureFromUUID(snapshot, idx+1)
		defendGestures = append(defendGestures, d)
		if result := models.PK(g, d); result > 0 {
			reward = reward.Add(reward)
		} else if result < 0 {
			reward = decimal.Zero
			break
		}
	}

	if reward.IsZero() {
		r.Result = models.Lose
	} else if reward.Equal(amount) {
		r.Result = models.Draw
	} else {
		r.Result = models.Win
	}

	r.Reward = reward.String()
	r.Gestures = models.GestureString(gestures)
	r.DefendGestures = models.GestureString(defendGestures)

	return r, nil
}

func refundSnapshot(ctx context.Context, s *Snapshot, memo string) error {
	traceId := traceIdFromSnapshotId(s.SnapshotId)
	_, err := models.CreatePayment(ctx, traceId, s.Asset.AssetId, s.Amount, s.OpponentId, memo)
	return err
}

func rewardRecord(ctx context.Context, r *models.Record, memo string) error {
	traceId := traceIdFromSnapshotId(r.SnapshotId)
	_, err := models.CreatePayment(ctx, traceId, r.AssetId, r.Reward, r.UserId, memo)
	return err
}
