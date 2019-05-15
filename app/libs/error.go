package libs

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris/context"
	"strconv"
	"strings"
)

type ErrorRes struct {
	Message string
}

func Assert(condition bool, msg string, code ...int) {
	if !condition {
		statusCode := 400
		if len(code) > 0 {
			statusCode = code[0]
		}
		panic("knownError&" + strconv.Itoa(statusCode) + "&" + msg)
	}
}

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
						b, errJson := jsoniter.Marshal(ErrorRes{
							Message: p[2],
						})
						if errJson != nil {
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
