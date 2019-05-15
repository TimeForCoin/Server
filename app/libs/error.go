package libs

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	"runtime"
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
						ctx.ContentType("application/javascript")
						_, err = ctx.Write(b)
						if err == nil && statusCode < 500 {
							return
						}
					}

				}

				var stacktrace string
				for i := 1; ; i++ {
					_, f, l, got := runtime.Caller(i)
					if !got {
						break
					}

					stacktrace += fmt.Sprintf("%s:%d\n", f, l)
				}

				// when stack finishes
				logMessage := fmt.Sprintf("Recovered from a route's Handler('%s')\n", ctx.HandlerName())
				logMessage += fmt.Sprintf("At Request: %s\n",
					fmt.Sprintf("%v %s %s %s", strconv.Itoa(ctx.GetStatusCode()), ctx.RemoteAddr(), ctx.Method(), ctx.Path()))
				logMessage += fmt.Sprintf("Trace: %s\n", err)
				logMessage += fmt.Sprintf("\n%s", stacktrace)
				ctx.Application().Logger().Warn(logMessage)

				ctx.StatusCode(iris.StatusInternalServerError)
				ctx.StopExecution()
			}
		}()

		ctx.Next()
	}
}
