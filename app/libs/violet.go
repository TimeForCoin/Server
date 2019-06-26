package libs

import (
	"github.com/TimeForCoin/Server/app/utils"
	"github.com/rs/zerolog/log"
	"gopkg.in/xmatrixstudio/violet.sdk.go.v3"
)

var oauth *OAuthService

// OAuthService Violet 授权服务
type OAuthService struct {
	API      *violet.Violet
	Callback string
}

// InitViolet 初始化授权服务
func InitViolet(c utils.VioletConfig) *OAuthService {
	oauth = &OAuthService{
		API: violet.NewViolet(violet.Config{
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
