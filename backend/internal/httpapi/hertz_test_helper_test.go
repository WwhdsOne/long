package httpapi

import (
	"strings"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/route"
)

func newNativeTestHandler(options Options) *nativeTestHandler {
	return &nativeTestHandler{engine: NewHertzServer(":0", options).Engine}
}

type nativeTestHandler struct {
	engine *route.Engine
}

func performJSONRequest(t *testing.T, handler *nativeTestHandler, method string, path string, body string, headers ...ut.Header) *protocol.Response {
	t.Helper()

	var requestBody *ut.Body
	if body != "" {
		requestBody = &ut.Body{
			Body: strings.NewReader(body),
			Len:  len(body),
		}
		headers = append(headers, ut.Header{Key: "Content-Type", Value: "application/json"})
	}

	recorder := ut.PerformRequest(handler.engine, method, path, requestBody, headers...)
	return recorder.Result()
}

func decodeJSONResponse[T any](t *testing.T, response *protocol.Response) T {
	t.Helper()

	var payload T
	if err := sonic.Unmarshal(response.Body(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}
