package libs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"gopkg.in/resty.v1"
)

var wechat *WeChatService

// WeChatService 微信小程序服务
type WeChatService struct {
	AppID        string
	AppSecret    string
	AppToken     string
	TokenExpires int64
}

// InitWeChat 初始化微信服务
func InitWeChat(c WechatConfig) {
	wechat = &WeChatService{
		AppID:        c.AppID,
		AppSecret:    c.AppSecret,
		AppToken:     "",
		TokenExpires: 0,
	}
}

// WeChatError 微信错误码
type WeChatError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// WeChatTokenRes 微信 Token 数据
type WeChatTokenRes struct {
	WeChatError
	AccessToken string `json:"access_token"`
	Expires     int64  `json:"expires_in"`
}

// getToken 获取应用 Token
func (s *WeChatService) getToken() (string, error) {
	// Token 已过期
	if time.Unix(s.TokenExpires, 0).Before(time.Now()) {
		resp, err := resty.R().
			SetQueryParam("grant_type", "client_credential").
			SetQueryParam("appid", s.AppID).
			SetQueryParam("secret", s.AppSecret).
			Get("https://api.weixin.qq.com/cgi-bin/token")
		if err != nil {
			return "", err
		}
		res := WeChatTokenRes{}
		if err := jsoniter.Unmarshal(resp.Body(), &res); err != nil {
			return "", err
		}
		if res.ErrCode != 0 {
			return "", errors.New(res.ErrMsg)
		}
		s.AppToken = res.AccessToken
		s.TokenExpires = time.Now().Unix() + res.Expires
	}
	return s.AppToken, nil
}

// WeChatMakeImageReq 生成微信小程序码请求
type WeChatMakeImageReq struct {
	Scene string `json:"scene"`
}

// MakeImage 生成小程序码
func (s *WeChatService) MakeImage(data string) (string, error) {
	token, err := s.getToken()
	if err != nil {
		return "", err
	}
	resp, err := resty.R().SetQueryParam("access_token", token).SetBody(WeChatMakeImageReq{
		Scene: data,
	}).Post("https://api.weixin.qq.com/wxa/getwxacodeunlimit")
	encodeString := base64.StdEncoding.EncodeToString(resp.Body())
	return "data:image/jpg;base64," + encodeString, nil
}

// WeChatSessionRes 微信登陆数据
type WeChatSessionRes struct {
	WeChatError
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
}

// GetOpenID 获取用户 OpenID
func (s *WeChatService) GetOpenID(code string) (string, error) {
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

// GetWeChat 获取微信服务
func GetWeChat() *WeChatService {
	if wechat == nil {
		log.Panic().Msg("Wechat service is not init")
	}
	return wechat
}
