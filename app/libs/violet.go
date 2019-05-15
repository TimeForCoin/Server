package libs

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/xmatrixstudio/violet.sdk.go.v3"
)

var oauth *OAuthService

type OAuthService struct {
	Api *violet.Violet
	Callback string
}

func InitViolet(c VioletConfig) *OAuthService {
	oauth = &OAuthService{
		Api: violet.NewViolet(violet.Config{
			ClientID: c.ClientID,
			ClientKey: c.ClientKey,
			ServerHost: c.ServerHost,
		}),
		Callback: c.Callback,
	}
	return oauth
}

func GetOauth () *OAuthService {
	if oauth == nil {
		log.Panic().Msg("OAuth service is not init")
	}
	return oauth
}