package libs

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// JSON 格式化结构体为 JSON 并写入 Response Body
// 自动将 Golang 中的驼峰命名法转换为 下划线命名法
func JSON(ctx context.Context, v interface{}) {
	b, err := jsoniter.Marshal(v)
	Assert(err == nil, "Error", iris.StatusInternalServerError)
	ctx.ContentType("application/json")
	_, err = ctx.Write(b)
	Assert(err == nil, "Error", iris.StatusInternalServerError)
}
