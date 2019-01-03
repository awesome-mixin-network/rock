package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/soooooooon/rock/commands"
	"github.com/soooooooon/rock/engine"
	"github.com/soooooooon/rock/models"
	"github.com/soooooooon/rock/server"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Rock"
	app.Version = "0.01"

	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug"},
	}

	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			log.SetLevel(log.DebugLevel)
			gin.SetMode(gin.DebugMode)
		}

		return nil
	}

	app.ExitErrHandler = func(c *cli.Context, err error) {
		if err != nil {
			log.Error(err)
		}
	}

	ctx := context.Background()

	app.Commands = append(app.Commands, cli.Command{
		Name:    "engine",
		Aliases: []string{"e"},
		Usage:   "launch rock engine",
		Flags:   []cli.Flag{cli.BoolFlag{Name: "now, n"}},
		Action: func(c *cli.Context) error {
			ctx := withSession(ctx)
			fromNow := c.Bool("now")
			return engine.LaunchEngine(ctx, fromNow)
		},
	})

	app.Commands = append(app.Commands, cli.Command{
		Name:    "server",
		Aliases: []string{"s"},
		Flags:   []cli.Flag{cli.IntFlag{Name: "port, p", Value: 8081}},
		Usage:   "launch rocker api server",
		Action: func(c *cli.Context) error {
			ctx := withSession(ctx)
			return server.LaunchServer(ctx, c.Int("port"))
		},
	})

	app.Commands = append(app.Commands, cli.Command{
		Name: "setdb",
		Action: func(c *cli.Context) error {
			ctx := withMysql(ctx)
			return models.SetDb(ctx)
		},
	})

	app.Commands = append(app.Commands, cli.Command{
		Name:  "rank",
		Usage: "update rank of all active arenas",
		Action: func(c *cli.Context) error {
			ctx := withMysql(ctx)
			return engine.UpdateArenaRanks(ctx)
		},
	})

	app.Commands = append(app.Commands,
		commands.CreateArena(ctx),
		commands.BetArena(ctx),
		commands.Deposit(ctx),
		commands.ReadAssets(ctx),
		commands.Login(ctx),
	)

	app.Run(os.Args)
}
