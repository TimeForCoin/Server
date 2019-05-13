package libs

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/xmatrixstudio/violet.sdk.go.v3"
)

var oauth *violet.Violet

func InitViolet(c VioletConfig) {
	oauth = violet.NewViolet(violet.Config{
		ClientID: c.ClientID,
		ClientKey: c.ClientKey,
		ServerHost: c.ServerHost,
	})
}

func GetOauth () *violet.Violet {
	if oauth == nil {
		log.Fatal().Msg("Violet is not init")
	}
	return oauth
}