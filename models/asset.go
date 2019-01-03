package models

import (
	"context"

	jsoniter "github.com/json-iterator/go"
	"github.com/soooooooon/rock/session"
)

const (
	allAssetKey = "rock_all_assets"
)

type Asset struct {
	AssetId string `json:"asset_id"`
	ChainId string `json:"chain_id"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
	Logo    string `json:"icon_url"`
}

func (a *Asset) Cache(ctx context.Context) error {
	data, err := jsoniter.Marshal(a)
	if err != nil {
		return err
	}

	_, err = session.Redis(ctx).HSet(allAssetKey, a.AssetId, data).Result()
	return err
}

func AssetFromCache(ctx context.Context, assetId string) (*Asset, error) {
	data, err := session.Redis(ctx).HGet(allAssetKey, assetId).Bytes()
	if err != nil {
		return nil, err
	}

	a := &Asset{}
	err = jsoniter.Unmarshal(data, a)
	return a, err
}

func AllAssetsFromCache(ctx context.Context) ([]*Asset, error) {
	data, err := session.Redis(ctx).HGetAll(allAssetKey).Result()
	if err != nil {
		return nil, err
	}

	assets := make([]*Asset, 0, len(data))
	for _, v := range data {
		a := &Asset{}
		if err := jsoniter.UnmarshalFromString(v, a); err == nil {
			assets = append(assets, a)
		}
	}

	return assets, nil
}
