package utils

import (
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris/context"
)

// ErrorRes 错误回复
type ErrorRes struct {
	Message string
}

// Assert 条件断言
// 当断言条件为 假 时触发 panic
// 对于当前请求不会再执行接下来的代码，并且返回指定格式的错误信息和错误码
func Assert(condition bool, msg string, code ...int) {
	if !condition {
		statusCode := 400
		if len(code) > 0 {
			statusCode = code[0]
		}
		panic("knownError&" + strconv.Itoa(statusCode) + "&" + msg)
	}
}

// AssertErr 错误断言
// 当 error 不为 nil 时触发 panic
// 对于当前请求不会再执行接下来的代码，并且返回指定格式的错误信息和错误码
// 若 msg 为空，则默认为 error 中的内容
func AssertErr(err error, msg string, code ...int) {
	if err != nil {
		statusCode := 400
		if len(code) > 0 {
			statusCode = code[0]
		}
		if msg == "" {
			msg = err.Error()
		}
		panic("knownError&" + strconv.Itoa(statusCode) + "&" + msg)
	}
}

// NewErrorHandler 错误捕获处理 Handler
func NewErrorHandler() context.Handler {
	return func(ctx context.Context) {
		defer func() {
			if err := recover(); err != nil {
				if ctx.IsStopped() {
					return
				}

				switch errStr := err.(type) {
				case string:
					p := strings.Split(errStr, "&")
					if len(p) == 3 && p[0] == "knownError" {
						statusCode, e := strconv.Atoi(p[1])
						if e != nil {
							break
						}
						ctx.StatusCode(statusCode)
						b, errJSON := jsoniter.Marshal(ErrorRes{
							Message: p[2],
						})
						if errJSON != nil {
							break
						}
						ctx.ContentType("application/json")
						_, err = ctx.Write(b)
						if err == nil && statusCode < 500 {
							return
						}
					}
				}
				panic(err)
			}
		}()
		ctx.Next()
	}
}
