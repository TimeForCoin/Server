package libs

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// JSON marshals the given interface object and writes the JSON response to the client.
func JSON(ctx context.Context, v interface{}) {
	b, err := jsoniter.Marshal(v)
	Assert(err == nil, "Error", iris.StatusInternalServerError)
	ctx.ContentType("application/json")
	_, err = ctx.Write(b)
	Assert(err == nil, "Error", iris.StatusInternalServerError)
}
