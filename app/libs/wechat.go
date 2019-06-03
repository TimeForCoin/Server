
package libs

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/resty.v1"
)

var wechat *WechatService

type WechatService struct {
	AppID string
	AppSecret string
}

func InitWeChat(c WechatConfig) {
	wechat = &WechatService{
		AppID: c.AppID,
		AppSecret: c.AppSecret,
	}
}


type WeChatSessionRes struct {
	OpenID string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID string `json:"unionid"`
	ErrCode int `json:"errcode"`
	ErrMsg string `json:"errmsg"`
}

func (s *WechatService) GetOpenID(code string) (string, error) {
	resp, err := resty.R().
		SetQueryParam("appid", s.AppID).
		SetQueryParam("secret", s.AppSecret).
		SetQueryParam("js_code", code).
		SetQueryParam("grant_type", "authorization_code").
		Post("https://api.weixin.qq.com/sns/jscode2session")
	if err != nil {
		return "", errors.New("network_error")
	}
	res := WeChatSessionRes{}
	if err = json.Unmarshal([]byte(resp.String()), &res); err != nil {
		return "", errors.New("wechat_error")
	}
	if res.ErrCode != 0 {
		return "", errors.New("error_code")
	}
	return res.OpenID, nil
}


func GetWechat () *WechatService {
	if oauth == nil {
		log.Panic().Msg("Wechat service is not init")
	}
	return wechat
}