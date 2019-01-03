package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/soooooooon/rock/config"
	"github.com/soooooooon/rock/models"
)

type Snapshot struct {
	SnapshotId string       `json:"snapshot_id"`
	Amount     string       `json:"amount"`
	Asset      models.Asset `json:"asset"`
	CreatedAt  time.Time    `json:"created_at"`

	TraceId    string `json:"trace_id"`
	UserId     string `json:"user_id"`
	OpponentId string `json:"opponent_id"`
	Data       string `json:"data"`
}

func requestMixinNetwork(ctx context.Context, checkpoint time.Time, limit int) ([]*Snapshot, error) {
	uri := fmt.Sprintf("/network/snapshots?offset=%s&order=ASC&limit=%d", checkpoint.Format(time.RFC3339Nano), limit)
	token, err := bot.SignAuthenticationToken(config.MixinClientId, config.MixinSessionId, config.MixinPrivateKey, "GET", uri, "")
	if err != nil {
		return nil, err
	}
	body, err := bot.Request(ctx, "GET", uri, nil, token)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data  []*Snapshot `json:"data"`
		Error string      `json:"error"`
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return resp.Data, nil
}

func ReadAssets(ctx context.Context) ([]bot.Asset, error) {
	token, err := bot.SignAuthenticationToken(config.MixinClientId, config.MixinSessionId, config.MixinPrivateKey, "GET", "/assets", "")
	if err != nil {
		return nil, fmt.Errorf("sign auth token failed: %s", err)
	}

	assets, err := bot.AssetList(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("read assets failed: %s", err)
	}

	return assets, nil
}
