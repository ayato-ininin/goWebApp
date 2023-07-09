package main

import (
	"context"
	"go_test_prac/webApp/pkg/data"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func Test_app_authenticate(t *testing.T) {
	theTests := []struct {
		name string
		requestBody string
		expectedStatusCode int
	}{
		{"valid user", `{"email":"admin@example.com","password":"secret"}`, http.StatusOK},
		{"not json", `I am not json`, http.StatusUnauthorized},
		{"empty json", `{}`, http.StatusUnauthorized},
		{"empty email", `{"email":"","password":"secret"}`, http.StatusUnauthorized},
		{"empty password", `{"email":"admin@example.com","password":""}`, http.StatusUnauthorized},
		{"invalid user", `{"email":"invalid@email.com","password":"test"}`, http.StatusUnauthorized},
	}

	for _, e := range theTests {
		var reader io.Reader
		reader = strings.NewReader(e.requestBody)
		req, _ := http.NewRequest("POST", "/auth", reader)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.authenticate)

		handler.ServeHTTP(rr, req)
		if e.expectedStatusCode != rr.Code {
			t.Errorf("test %s, expected status %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}

func Test_app_refresh(t *testing.T) {
	var tests = []struct {
		name string
		token string
		expectedStatusCode int
		resetRefreshTime bool
	}{
		{name: "valid", token: "", expectedStatusCode: http.StatusOK, resetRefreshTime: true},
		{name: "valid but not yet ready to expire", token: "", expectedStatusCode: http.StatusTooEarly, resetRefreshTime: false},
		{name: "expired token", token: expiredToken, expectedStatusCode: http.StatusBadRequest, resetRefreshTime: false},
	}

	testUser:= data.User {
		ID: 1,
		FirstName: "Admin",
		LastName: "User",
		Email: "admin@example.com",
	}

	oldRefreshTime := refreshTokenExpiry

	for _, e := range tests {
		var tkn string
		if e.token == "" {
			if e.resetRefreshTime {
				refreshTokenExpiry = time.Second * 1 // 1秒後に期限切れになる、テスト用
			}
			tokens, _ := app.generateTokenPair(&testUser)
			tkn = tokens.RefreshToken
		} else {
			tkn = e.token
		}

		postedData := url.Values{
			"refresh_token": {tkn},
		}

		req, _ := http.NewRequest("POST", "/refresh", strings.NewReader(postedData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.refresh)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: expected status %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		refreshTokenExpiry = oldRefreshTime
	}
}

func Test_app_userHandlers(t *testing.T) {
	var tests = []struct {
		name string
		method string
		json string
		paramID string
		handler http.HandlerFunc
		expectedStatusCode int
	}{
		{name:"allUsers", method:"GET", json:"", paramID:"", handler:app.allUsers, expectedStatusCode:http.StatusOK},
		{name:"deleteUser", method:"DELETE", json:"", paramID:"1", handler:app.deleteUser, expectedStatusCode:http.StatusNoContent},
		{name:"getUser valid"	, method:"GET", json:"", paramID:"1", handler:app.getUser, expectedStatusCode:http.StatusOK},
		{name:"getUser invalid", method:"GET", json:"", paramID:"2", handler:app.getUser, expectedStatusCode:http.StatusInternalServerError},
	}

	for _, e := range tests {
		var req *http.Request
		if e.json != "" {
			req, _ = http.NewRequest(e.method, "/", strings.NewReader(e.json))
		} else {
			req, _ = http.NewRequest(e.method, "/", nil)
		}

		if e.paramID != "" {
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", e.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(e.handler)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: expected status %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}
