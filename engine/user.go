package engine

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	log "github.com/sirupsen/logrus"
	"github.com/soooooooon/rock/config"
	"github.com/soooooooon/rock/models"
	"github.com/soooooooon/rock/session"
)

type MixinMessagerUser struct {
	MixinId   string    `json:"user_id"`
	Name      string    `json:"full_name"`
	Avatar    string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

func ReadUser(ctx context.Context, mixinId string) (*MixinMessagerUser, error) {
	uri := "/users/" + mixinId
	token, err := bot.SignAuthenticationToken(config.MixinClientId, config.MixinSessionId, config.MixinPrivateKey, "GET", uri, "")
	if err != nil {
		return nil, err
	}
	body, err := bot.Request(ctx, "GET", uri, nil, token)
	if err != nil {
		return nil, err
	}
	log.Debugf("read user: %s", string(body))
	var resp struct {
		Data  *MixinMessagerUser `json:"data"`
		Error string             `json:"error"`
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

func handleUpdateUsers(ctx context.Context) error {
	users, err := models.ListOutdatedUsers(ctx, 10)
	if err != nil {
		return err
	}

	for _, user := range users {
		mmu, err := ReadUser(ctx, user.MixinID)
		if err != nil {
			log.Errorf("read mixin messager user failed: %s", err)
			continue
		}

		updates := map[string]interface{}{
			"refreshed_at": time.Now().UTC(),
		}

		if len(mmu.Name) > 0 {
			updates["name"] = mmu.Name
			updates["avatar"] = mmu.Avatar
		}

		if err := session.MysqlWrite(ctx).Model(user).Updates(updates).Error; err != nil {
			log.Errorf("update user failed: %s", err)
			continue
		}

		user.Cache(ctx)
	}

	if len(users) >= limit {
		return handleUpdateUsers(ctx)
	}

	return nil
}
