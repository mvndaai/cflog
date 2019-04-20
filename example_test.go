package cflog_test

import (
	"context"

	"github.com/mvndaai/cflog"
)

var ctx = context.Background()

func Example_string() {
	cflog.Debug(ctx, "string")
}

func Example_jsonString() {
	cflog.Warn(ctx, `{"message": "json string"}`)
}

func Example_jsonStruct() {
	type s struct {
		Message string `json:"message"`
	}

	cflog.Error(ctx, s{Message: "json struct"})
}

func ExampleNewClient() {
	c, err := cflog.NewClient(context.Background())
	if err != nil {
		//...
	}
	defer c.Close()
	if err := c.Log(context.Background(), cflog.SeverityDebug, "Debug message"); err != nil {
		//...
	}
}
