package libs

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/xmatrixstudio/violet.sdk.go.v3"
)

var oauth *OAuthService

// OAuthService Violet 授权服务
type OAuthService struct {
	Api      *violet.Violet
	Callback string
}

// InitViolet 初始化授权服务
func InitViolet(c VioletConfig) *OAuthService {
	oauth = &OAuthService{
		Api: violet.NewViolet(violet.Config{
			ClientID:   c.ClientID,
			ClientKey:  c.ClientKey,
			ServerHost: c.ServerHost,
		}),
		Callback: c.Callback,
	}
	return oauth
}

// GetOAuth 获取授权服务
func GetOAuth() *OAuthService {
	if oauth == nil {
		log.Panic().Msg("OAuth service is not init")
	}
	return oauth
}
