package libs

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/rs/zerolog/log"
	"github.com/tencentyun/cos-go-sdk-v5"
	"gopkg.in/resty.v1"
	"mime/multipart"
	"net/http"
	"net/url"
)

var cosService *COSService

// COSService 对象存储服务
type COSService struct {
	Client *cos.Client
	URL    string
}

// InitCOS 初始化对象存储
func InitCOS(c COSConfig) {
	u, _ := url.Parse(c.URL)
	b := &cos.BaseURL{BucketURL: u}
	cosService = &COSService{
		Client: cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  c.AppID,
				SecretKey: c.AppSecret,
			},
		}),
		URL: c.URL,
	}
}

// GetCOS 获取对象存储实例
func GetCOS() *COSService {
	if cosService == nil {
		log.Panic().Msg("Cos service is not init")
	}
	return cosService
}

// DeleteFile 删除文件
func (s *COSService) DeleteFile(name string) error {
	_, err := s.Client.Object.Delete(context.Background(), name)
	return err
}

// SaveFile 保存文件
func (s *COSService) SaveFile(name string, file multipart.File) (url string, err error) {
	_, err = s.Client.Object.Put(context.Background(), name, file, nil)
	url = s.URL + "/" + name
	return
}

// SaveByteFile 保存二进制文件数据
func (s *COSService) SaveByteFile(name string, data []byte) (url string, err error) {
	r := bytes.NewReader(data)
	_, err = s.Client.Object.Put(context.Background(), name, r, nil)
	url = s.URL + "/" + name
	return
}

// SaveURLFile 保存链接文件数据
func (s *COSService) SaveURLFile(name string, fileURL string) (url string, err error) {
	resp, err := resty.R().Get(fileURL)
	return s.SaveByteFile(name, resp.Body())
}

// SaveBase64File 保存 Base64 文件
func (s *COSService) SaveBase64File(name, base string) (url string, err error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(base)
	if err != nil {
		return "", err
	}
	return s.SaveByteFile(name, decodeBytes)
}
