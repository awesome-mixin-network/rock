package commands

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/soooooooon/rock/config"
	"github.com/soooooooon/rock/memo"
	"github.com/soooooooon/rock/models"
	"github.com/urfave/cli"
)

func payUri(clientId, asset, amount, data, traceId string) string {
	u := url.URL{}
	u.Scheme = "mixin"
	u.Path = "pay"

	q := u.Query()
	q.Add("recipient", clientId)
	q.Add("asset", asset)
	q.Add("amount", amount)
	q.Add("memo", data)
	q.Add("trace", traceId)

	u.RawQuery = q.Encode()
	return u.String()
}

func CreateArena(ctx context.Context) cli.Command {
	command := cli.Command{
		Name:  "arena",
		Usage: "display a qrcode for creating new arena",
	}

	command.Flags = []cli.Flag{
		cli.StringFlag{Name: "symbol, s"},
		cli.StringFlag{Name: "amount, a"},
		cli.StringFlag{Name: "max, m"},
		cli.Int64Flag{Name: "expire, e", Value: 1},
		cli.StringFlag{Name: "trace, t"},
	}

	command.Action = func(c *cli.Context) error {
		symbol := strings.ToUpper(c.String("symbol"))
		assetId, ok := config.MarketId(symbol)
		if !ok {
			return fmt.Errorf("%s is not supported yet", symbol)
		}

		amount, _ := decimal.NewFromString(c.String("amount"))
		if !amount.IsPositive() {
			return fmt.Errorf("amount must be positive")
		}

		max, _ := decimal.NewFromString(c.String("max"))
		if !max.IsPositive() {
			max = amount.Shift(-2)
		} else if max.GreaterThan(amount) {
			max = amount
		}

		expire := c.Int64("expire")
		if expire < 1 {
			return fmt.Errorf("expire must be positive")
		}

		action := memo.Arena{
			E: expire,
			M: max.String(),
		}

		m, err := memo.Marshal(action)
		if err != nil {
			return err
		}

		trace := c.String("trace")
		if len(trace) == 0 {
			trace = uuid.Must(uuid.NewV4()).String()
		}

		uri := payUri(config.MixinClientId, assetId, amount.String(), m, trace)
		displayQRCode(uri)

		return nil
	}

	return command
}

func BetArena(ctx context.Context) cli.Command {
	command := cli.Command{
		Name:  "bet",
		Usage: "display a qrcode for beating an arena",
	}

	command.Flags = []cli.Flag{
		cli.StringFlag{Name: "symbol, s"},
		cli.StringFlag{Name: "amount, a"},
		cli.IntSliceFlag{Name: "gesture, g"},
		cli.StringFlag{Name: "id, i"},
		cli.StringFlag{Name: "trace, t"},
	}

	command.Action = func(c *cli.Context) error {
		symbol := strings.ToUpper(c.String("symbol"))
		assetId, ok := config.MarketId(symbol)
		if !ok {
			return fmt.Errorf("%s is not supported yet", symbol)
		}

		amount, _ := decimal.NewFromString(c.String("amount"))
		if !amount.IsPositive() {
			return fmt.Errorf("amount must be positive")
		}

		gestures := c.IntSlice("gesture")
		if len(gestures) == 0 {
			return fmt.Errorf("no gestures")
		}

		id := c.String("id")
		if len(id) == 0 {
			return fmt.Errorf("no arena id")
		}

		trace := c.String("trace")
		if len(trace) == 0 {
			trace = uuid.Must(uuid.NewV4()).String()
		}

		action := memo.Bet{
			A: id,
			G: models.GestureString(gestures),
		}

		m, err := memo.Marshal(action)
		if err != nil {
			return err
		}

		uri := payUri(config.MixinClientId, assetId, amount.String(), m, trace)
		displayQRCode(uri)

		return nil
	}

	return command
}

func Login(ctx context.Context) cli.Command {
	command := cli.Command{
		Name:  "login",
		Usage: "display a QRCode for logging",
	}

	command.Flags = []cli.Flag{
		cli.StringFlag{Name: "symbol, s", Value: "CNB"},
		cli.StringFlag{Name: "trace, t"},
	}

	command.Action = func(c *cli.Context) error {
		symbol := c.String("symbol")
		assetId, ok := config.MarketId(symbol)
		if !ok {
			return fmt.Errorf("%s is not supported yet", symbol)
		}

		action := memo.Login{
			T: time.Now().Unix(),
		}
		data, err := memo.Marshal(action)
		if err != nil {
			return err
		}

		trace := c.String("trace")
		if len(trace) == 0 {
			trace = uuid.Must(uuid.NewV4()).String()
		}

		uri := payUri(config.MixinClientId, assetId, "1", data, trace)
		displayQRCode(uri)
		fmt.Printf("\ntoken = %s\n", trace)
		return nil
	}

	return command
}
