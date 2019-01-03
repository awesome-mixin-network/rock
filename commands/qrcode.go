package commands

import (
	"os"

	"github.com/mdp/qrterminal"
	log "github.com/sirupsen/logrus"
)

func displayQRCode(u string) {
	log.Debugf("QRCode : %s", u)
	qrterminal.GenerateHalfBlock(u, qrterminal.H, os.Stdout)
}
