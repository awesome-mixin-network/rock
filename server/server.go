package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/soooooooon/gin-contrib/gin_helper"
	"github.com/soooooooon/gin-contrib/limiter"
	"github.com/soooooooon/rock/models"
	"github.com/soooooooon/rock/session"
)

func LaunchServer(ctx context.Context, port int) error {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	corsOp := cors.DefaultConfig()
	corsOp.AllowCredentials = true
	corsOp.AllowHeaders = []string{"Authorization", "Origin", "Content-Type"}
	corsOp.AllowOriginFunc = func(origin string) bool {
		return true
	}
	r.Use(cors.New(corsOp))
	r.Use(func(c *gin.Context) { c.Request = c.Request.WithContext(ctx) })

	redisClient := session.Redis(ctx)
	if limit, err := limiter.NewLimiterWithClient(redisClient); err == nil {
		limit.AddGroup("READ", 512, time.Minute)
		r.Use(limit.Limit())
	} else {
		log.Errorf("create new limiter failed: %s", err)
		return err
	}

	api := r.Group("/api")
	login := r.Group("/api", LoginRequired)

	weight := func(w int) limiter.WeightFunc {
		return func(c *gin.Context) int { return w }
	}

	api.GET("/_hc", func(c *gin.Context) { gin_helper.OK(c) })
	api.GET("/_status", handleServerStatus())
	api.GET("config", handleQueryConfig)
	// arenas and records
	api.GET("/arena/:id", limiter.Available("READ", weight(1)), handleArenaDetail)
	api.GET("/arenas/new", limiter.Available("READ", weight(1)), handleQueryNewArenas)
	api.GET("/arenas/top", limiter.Available("READ", weight(1)), handleExploreArenas)
	api.GET("/record/:id", limiter.Available("READ", weight(1)), handleRecordDetail)
	api.GET("/arena/:id/records", limiter.Available("READ", weight(1)), handleQueryRecords)
	api.GET("/assets", limiter.Available("READ", weight(1)), handleQueryAssets)

	// ws
	api.GET("/ws", handleWebsocketRequests(ctx))

	login.GET("/me", limiter.Available("READ", weight(1)), queryMyProfile)
	login.GET("/arenas/my", limiter.Available("READ", weight(1)), handleQueryMyArenas)
	login.GET("/records/my", limiter.Available("READ", weight(1)), handleQueryMyRecords)

	r.NoRoute(func(c *gin.Context) { gin_helper.FailError(c, pageNotFound) })

	addr := fmt.Sprintf(":%d", port)
	log.Debugf("launch server at %s", addr)
	return r.Run(addr)
}

func handleServerStatus() gin.HandlerFunc {
	loc, _ := time.LoadLocation("Asia/Chongqing")
	start := time.Now().In(loc)
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		engineCheckpoint, err := models.ReadPropertyAsTime(ctx, "rock_bot_mixin_network")
		if err != nil {
			gin_helper.FailError(c, internalErr, err.Error())
			return
		}

		gin_helper.OK(c, "server", start, "engine", engineCheckpoint.In(loc))
	}
}
