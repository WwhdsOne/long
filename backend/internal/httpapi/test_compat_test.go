package httpapi

import (
	"bytes"
	"io"
	"net/http"

	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"
)

// NewHandler 仅供现有测试复用，底层已切换为 Hertz 原生路由。
func NewHandler(options Options) *testHandler {
	return &testHandler{engine: newTestEngine(options)}
}

type testHandler struct {
	engine *route.Engine
}

func newTestEngine(options Options) *route.Engine {
	return NewHertzServer(":0", options).Engine
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body *ut.Body
	if r.Body != nil {
		payload, _ := io.ReadAll(r.Body)
		body = &ut.Body{
			Body: bytes.NewReader(payload),
			Len:  len(payload),
		}
	}

	headers := make([]ut.Header, 0, len(r.Header))
	for key, values := range r.Header {
		for _, value := range values {
			headers = append(headers, ut.Header{Key: key, Value: value})
		}
	}

	resp := ut.PerformRequest(h.engine, r.Method, r.URL.RequestURI(), body, headers...).Result()
	resp.Header.VisitAll(func(key, value []byte) {
		w.Header().Add(string(key), string(value))
	})
	w.WriteHeader(resp.StatusCode())
	_, _ = w.Write(resp.Body())
}
