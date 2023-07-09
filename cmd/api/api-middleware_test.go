package main

import (
	"fmt"
	"go_test_prac/webApp/pkg/data"
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


func Test_app_authRequired(t *testing.T) {
	// dummy handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	testUser := data.User {
		ID: 1,
		FirstName: "Admin",
		LastName: "User",
		Email: "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	var tests = []struct {
		name string
		token string
		expectAuthorized bool
		setHeader bool
	}{
		{name: "valid", token: fmt.Sprintf("Bearer %s", tokens.Token), expectAuthorized: true, setHeader: true},
		{name: "no token", token: "", expectAuthorized: false, setHeader: false},
		{name: "invalid token", token: fmt.Sprintf("Bearer %s", expiredToken), expectAuthorized: false, setHeader: true},
	}

	for _, e := range tests {
		req, _ := http.NewRequest("GET", "/", nil)
		if e.setHeader {
			req.Header.Set("Authorization", e.token)
		}
		rr := httptest.NewRecorder()
		handlerToTest := app.authRequired(nextHandler)
		handlerToTest.ServeHTTP(rr, req)

		if e.expectAuthorized && rr.Code == http.StatusUnauthorized {
			t.Errorf("%s: got code 401, and should not have", e.name)
		}

		if !e.expectAuthorized && rr.Code != http.StatusUnauthorized {
			t.Errorf("%s: did not get code 402, and should have", e.name)
		}
	}
}
