package libs

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
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