package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_application_addIPToContext(t *testing.T) {
	tests := []struct {
		headerName string // x-forwarded-for
		headerValue string // x-forwarded-for:value
		addr string // remoteAddress
		emptyAddr bool // remote addressが空かどうか
	}{
		{"", "", "", false}, // 空のremote address,req.RemoteAddrには、テスト用の192.0.2.1がそのまま入っている？
		{"","", "", true}, // そもそもremote addressがないので、192.0.2.1を""にして、unknown
		{"X-Forwarded-For", "192.3.2.1", "", false}, // x-forwarded-forを指定
		{"","","192.168.1.0:8000",false}, // remote addressを指定
	}

	var app application

	// create a dummy handler that we'll use to check the context
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//make sure that the value exitsts in the context
		val := r.Context().Value(contextUserKey)
		if val == nil {
			t.Error(contextUserKey, "not found in context")
		}

		// make sure we got a string
		ip, ok := val.(string)
		if !ok {
			t.Error("value is not a string")
		}
		t.Log(ip)
	})

	for _, e := range tests {
		// create the hander to test
		handlerToTest := app.appIPToContext(nextHandler)

		// mock request
		req := httptest.NewRequest("GET", "http://testing", nil)
		// emptyフラグがあれば空にして、""でも192.0.2.1が入っている
		if e.emptyAddr {
			req.RemoteAddr = ""
		}

		// x-forwarded-forがあればいれる
		if len(e.headerName) > 0 {
			req.Header.Add(e.headerName, e.headerValue)
		}

		// remote addressがあればいれる
		if len(e.addr) > 0 {
			req.RemoteAddr = e.addr
		}

		// 作成したリクエストを流す
		handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
	}
}
