package libs

import (
	"crypto/sha512"
	"encoding/hex"
	"math/rand"
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
