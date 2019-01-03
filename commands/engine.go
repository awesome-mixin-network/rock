package commands

import (
	"context"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/soooooooon/rock/config"
	"github.com/soooooooon/rock/engine"
	"github.com/urfave/cli"
)

func ReadAssets(ctx context.Context) cli.Command {
	command := cli.Command{
		Name:  "assets",
		Usage: "read engine's assets",
	}

	command.Action = func(c *cli.Context) error {
		assets, err := engine.ReadAssets(ctx)
		if err != nil {
			return err
		}

		var sumUSD, sumBTC decimal.Decimal
		for _, a := range assets {
			log.Infof("%-6s %s  %s", a.Symbol, a.AssetId, a.Balance)

			balance, _ := decimal.NewFromString(a.Balance)
			priceUSD, _ := decimal.NewFromString(a.PriceUSD)
			sumUSD = sumUSD.Add(balance.Mul(priceUSD))
			priceBTC, _ := decimal.NewFromString(a.PriceBTC)
			sumBTC = sumBTC.Add(balance.Mul(priceBTC))
		}

		log.Infof("SUM: %s USD, %s BTC", sumUSD.String(), sumBTC.String())
		return nil
	}

	return command
}

func Deposit(ctx context.Context) cli.Command {
	command := cli.Command{
		Name:  "deposit",
		Usage: "deposit to rock engine",
	}

	command.Action = func(c *cli.Context) error {
		uri := "mixin://transfer/" + config.MixinClientId
		displayQRCode(uri)
		return nil
	}

	return command
}
