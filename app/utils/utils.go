package utils

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"io"
	"math/rand"
	"mime/multipart"
	"time"
)

// GetRandomString 获取随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// GetHash 获取字符串的 Hash
func GetHash(data string) string {
	h := sha512.New()
	h.Write([]byte(data))
	md := h.Sum(nil)
	return hex.EncodeToString(md)
}

// GetFileHash 获取文件 Hash
func GetFileHash(file multipart.File) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	_, err := file.Seek(0, 0)
	if err != nil {
		return "", err
	}
	md := hash.Sum(nil)
	return hex.EncodeToString(md), nil
}

// JSON 格式化结构体为 JSON 并写入 Response Body
// 自动将 Golang 中的驼峰命名法转换为 下划线命名法
func JSON(ctx context.Context, v interface{}) {
	b, err := jsoniter.Marshal(v)
	Assert(err == nil, "Error", iris.StatusInternalServerError)
	ctx.ContentType("application/json")
	_, err = ctx.Write(b)
	Assert(err == nil, "Error", iris.StatusInternalServerError)
}
