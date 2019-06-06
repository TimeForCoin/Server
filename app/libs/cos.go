package libs

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/rs/zerolog/log"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
)

var cosService *COSService

type COSService struct {
	Client *cos.Client
	URL string
}

func InitCOS(c COSConfig)  {
	u, _ := url.Parse(c.URL)
	b := &cos.BaseURL{BucketURL: u}
	cosService = &COSService{
		Client: cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID: c.AppID,
				SecretKey: c.AppSecret,
			},
		}),
		URL: c.URL,
	}
}

func GetCOS () *COSService {
	if cosService == nil {
		log.Panic().Msg("Cos service is not init")
	}
	return cosService
}

func (s *COSService) SaveFile(name string, data []byte) (url string, err error) {
	r := bytes.NewReader(data)
	_, err = s.Client.Object.Put(context.Background(), name, r, nil)
	url = s.URL + "/" + name
	return
}

func (s *COSService) SaveBase64(name, base string) (url string, err error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(base)
	if err != nil {
		return "", err
	}
	return s.SaveFile(name, decodeBytes)
}
