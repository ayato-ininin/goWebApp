package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// このミドルウェアを通った場合に正しくヘッダーが設定されるかをテスト
// methodがOPTIONSの場合
func Test_app_enableCORS(t *testing.T) {
	// dummy handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	var tests = []struct {
		name string
		method string
		expectHeader bool
	}{
		{"preflight request", http.MethodOptions, true},
		{"get request", http.MethodGet, false},
	}

	for _, e := range tests {
		handlerToTest := app.enableCORS(nextHandler)

		req := httptest.NewRequest(e.method, "http://testing", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		if e.expectHeader && rr.Header().Get("Access-Control-Allow-Credentials") == "" {
			t.Errorf("%s: expected header Access-Control-Allow-Credentials, but not found", e.name)
		}

		if !e.expectHeader && rr.Header().Get("Access-Control-Allow-Credentials") != "" {
			t.Errorf("%s: unexpected header Access-Control-Allow-Credentials, but found", e.name)
		}
	}
}
