package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"github.com/soooooooon/gin-contrib/gin_helper"
	"github.com/soooooooon/rock/config"
	"github.com/soooooooon/rock/models"
)

type arenaView struct {
	*models.Arena
	ID    string        `json:"id"`
	Asset *models.Asset `json:"asset,omitempty"`
}

func handleArenaDetail(c *gin.Context) {
	id := c.Param("id")

	var (
		ctx               = c.Request.Context()
		a   *models.Arena = nil
	)

	if _, err := uuid.FromString(id); err == nil {
		a, _ = models.ArenaFromCache(ctx, id)
	} else if hashid, err := models.HashIdDecode(id); err == nil {
		a, _ = models.ArenaWithId(ctx, hashid)
	}

	if a == nil {
		gin_helper.FailError(c, pageNotFound)
		return
	}

	view := arenaView{
		Arena: a,
		ID:    models.HashIdEncode(a.ID),
	}

	if asset, err := models.AssetFromCache(ctx, a.AssetId); err == nil {
		view.Asset = asset
	}

	gin_helper.OK(c, view)
}

func handleQueryNewArenas(c *gin.Context) {
	form := struct {
		AssetId string `form:"asset_id"`
		Cursor  string `form:"cursor"`
		Limit   int    `form:"limit"`
	}{}

	gin_helper.BindQuery(c, &form)

	req := &models.ArenaRequest{
		AssetId:       form.AssetId,
		OnlyUnexpired: true,
	}

	if len(form.Cursor) > 0 {
		req.FromId, _ = models.HashIdDecode(form.Cursor)
	}

	limit := gin_helper.Limit(form.Limit, 50, 20)
	ctx := c.Request.Context()

	arenas, err := models.QueryArenas(ctx, req, limit+1)
	if err != nil {
		gin_helper.FailError(c, internalErr, err.Error())
		return
	}

	cursor := ""
	if len(arenas) > limit {
		cursor = models.HashIdEncode(arenas[limit-1].ID)
		arenas = arenas[:limit]
	}

	views := make([]arenaView, len(arenas))
	assets := map[string]*models.Asset{}

	for idx, a := range arenas {
		v := arenaView{
			Arena: a,
			ID:    models.HashIdEncode(a.ID),
		}

		if asset, ok := assets[a.AssetId]; ok {
			v.Asset = asset
		} else {
			asset, _ := models.AssetFromCache(ctx, a.AssetId)
			v.Asset = asset
			assets[a.AssetId] = asset
		}

		views[idx] = v
	}

	gin_helper.OkWithPagination(c, cursor, "arenas", views)
}

func handleExploreArenas(c *gin.Context) {
	ctx := c.Request.Context()
	arenas, err := models.ListTopRankArenas(ctx, 50)
	if err != nil {
		gin_helper.FailError(c, internalErr, err.Error())
		return
	}

	views := make([]arenaView, len(arenas))
	assets := map[string]*models.Asset{}

	for idx, a := range arenas {
		v := arenaView{
			Arena: a,
			ID:    models.HashIdEncode(a.ID),
		}

		if asset, ok := assets[a.AssetId]; ok {
			v.Asset = asset
		} else {
			asset, _ := models.AssetFromCache(ctx, a.AssetId)
			v.Asset = asset
			assets[a.AssetId] = asset
		}

		views[idx] = v
	}

	gin_helper.OK(c, views)
}

func handleQueryMyArenas(c *gin.Context) {
	form := struct {
		Cursor string `form:"cursor"`
		Limit  int    `form:"limit"`
	}{}
	gin_helper.BindQuery(c, &form)

	user := ExtractMixinUserId(c)
	req := &models.ArenaRequest{
		UserId: user,
	}

	if len(form.Cursor) > 0 {
		req.FromId, _ = models.HashIdDecode(form.Cursor)
	}
	limit := gin_helper.Limit(form.Limit, 50, 20)

	ctx := c.Request.Context()
	arenas, err := models.QueryArenas(ctx, req, limit)
	if err != nil {
		gin_helper.FailError(c, internalErr, err.Error())
		return
	}

	cursor := ""
	if len(arenas) > limit {
		cursor = models.HashIdEncode(arenas[limit-1].ID)
		arenas = arenas[:limit]
	}

	views := make([]arenaView, len(arenas))
	assets := map[string]*models.Asset{}

	for idx, a := range arenas {
		v := arenaView{
			Arena: a,
			ID:    models.HashIdEncode(a.ID),
		}

		if asset, ok := assets[a.AssetId]; ok {
			v.Asset = asset
		} else {
			asset, _ := models.AssetFromCache(ctx, a.AssetId)
			v.Asset = asset
			assets[a.AssetId] = asset
		}

		views[idx] = v
	}

	gin_helper.OkWithPagination(c, cursor, "arenas", views)
}

func handleRecordDetail(c *gin.Context) {
	id := c.Param("id")

	if _, err := uuid.FromString(id); err != nil {
		gin_helper.FailError(c, pageNotFound, "%s is not uuid", id)
		return
	}

	ctx := c.Request.Context()
	r, err := models.RecordFromCache(ctx, id)
	if err != nil {
		gin_helper.FailError(c, pageNotFound, err.Error())
		return
	}

	gin_helper.OK(c, r)
}

func handleQueryRecords(c *gin.Context) {
	id := c.Param("id")
	arenaId, err := models.HashIdDecode(id)
	if err != nil {
		gin_helper.FailError(c, fmt.Errorf("invalid arena id %s", id))
	}

	req := &models.RecordRequest{
		ArenaId: arenaId,
		Desc:    true,
	}

	form := struct {
		Cursor string `form:"cursor"`
		Limit  int    `form:"limit"`
	}{}

	gin_helper.BindQuery(c, &form)

	if cursor := c.Query("cursor"); len(cursor) > 0 {
		req.FromId, _ = models.HashIdDecode(cursor)
	}

	limit := gin_helper.Limit(form.Limit, 50, 20)
	ctx := c.Request.Context()

	records, err := models.QueryRecords(ctx, req, limit+1)
	if err != nil {
		gin_helper.FailError(c, internalErr, err.Error())
		return
	}

	cursor := ""
	if len(records) > limit {
		cursor = models.HashIdEncode(records[limit-1].ID)
		records = records[:limit]
	}

	gin_helper.OkWithPagination(c, cursor, "records", records)
}

func handleQueryMyRecords(c *gin.Context) {
	req := &models.RecordRequest{
		UserId: ExtractMixinUserId(c),
		Desc:   true,
	}

	form := struct {
		Cursor string `form:"cursor"`
		Limit  int    `form:"limit"`
	}{}

	gin_helper.BindQuery(c, &form)

	if cursor := c.Query("cursor"); len(cursor) > 0 {
		req.FromId, _ = models.HashIdDecode(cursor)
	}

	limit := gin_helper.Limit(form.Limit, 50, 20)
	ctx := c.Request.Context()

	records, err := models.QueryRecords(ctx, req, limit+1)
	if err != nil {
		gin_helper.FailError(c, internalErr, err.Error())
		return
	}

	cursor := ""
	if len(records) > limit {
		cursor = models.HashIdEncode(records[limit-1].ID)
		records = records[:limit]
	}

	gin_helper.OkWithPagination(c, cursor, "records", records)
}

func handleQueryConfig(c *gin.Context) {
	config := map[string]interface{}{
		"client_id": config.MixinClientId,
	}

	gin_helper.OK(c, config)
}

func handleQueryAssets(c *gin.Context) {
	ctx := c.Request.Context()
	if assets, err := models.AllAssetsFromCache(ctx); err == nil {
		gin_helper.OK(c, assets)
	} else {
		gin_helper.FailError(c, err)
	}
}
