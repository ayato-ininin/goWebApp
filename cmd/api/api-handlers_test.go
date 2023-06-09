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
		{name:"deleteUser bad URL param", method:"DELETE", json:"", paramID:"YYY", handler:app.deleteUser, expectedStatusCode:http.StatusBadRequest},
		{name:"getUser valid"	, method:"GET", json:"", paramID:"1", handler:app.getUser, expectedStatusCode:http.StatusOK},
		{name:"getUser invalid", method:"GET", json:"", paramID:"2", handler:app.getUser, expectedStatusCode:http.StatusInternalServerError},
		{name:"getUser bad URL param", method:"GET", json:"", paramID:"YYY", handler:app.getUser, expectedStatusCode:http.StatusBadRequest},
		{
			name:"updateUser valid",
			method:"PATCH",
			json:`{"id":1,"first_name":"admin","last_name":"user","email":"admin@example.com"}`,
			paramID:"",
			handler: app.updateUser,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:"updateUser invalid",
			method:"PATCH",
			json:`{"id":2,"first_name":"admin","last_name":"user","email":"admin@example.com"}`,
			paramID:"",
			handler: app.updateUser,
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name:"updateUser invalid json",
			method:"PATCH",
			json:`{"id":1,first_name:"admin","last_name":"user","email":"admin@example.com"}`,
			paramID:"",
			handler: app.updateUser,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:"insertUser valid",
			method:"PUT",
			json:`{"first_name":"jack","last_name":"test","email":"admin@example.com"}`,
			paramID:"",
			handler: app.insertUser,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:"insertUser invalid",
			method:"PUT",
			json:`{"foo":"bar","first_name":"jack","last_name":"test","email":"admin@example.com"}`,
			paramID:"",
			handler: app.insertUser,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:"insertUser invalid json",
			method:"PUT",
			json:`{first_name":"jack","last_name":"test","email":"admin@example.com"}`,
			paramID:"",
			handler: app.insertUser,
			expectedStatusCode: http.StatusBadRequest,
		},
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

func Test_app_refreshUsingCookie(t *testing.T) {
	testUser := data.User{
		ID: 1,
		FirstName: "Admin",
		LastName: "User",
		Email: "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	testCookie := http.Cookie{
		Name: "Host-refresh_token",
		Value: tokens.RefreshToken,
		Path: "/",
		Expires: time.Now().Add(refreshTokenExpiry),
		MaxAge: int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain: "localhost",
		HttpOnly: true,
		Secure: true,
	}
	badCookie := http.Cookie{
		Name: "Host-refresh_token",
		Value: "bad-token",
		Path: "/",
		Expires: time.Now().Add(refreshTokenExpiry),
		MaxAge: int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain: "localhost",
		HttpOnly: true,
		Secure: true,
	}

	var tests = []struct {
		name string
		addCookie bool
		cookie *http.Cookie
		expectedStatusCode int
	}{
		{name:"valid", addCookie:true, cookie:&testCookie, expectedStatusCode:http.StatusOK},
		{name:"invalid", addCookie:true, cookie:&badCookie, expectedStatusCode:http.StatusBadRequest},
		{name:"no cookie", addCookie:false, cookie:nil, expectedStatusCode:http.StatusUnauthorized},
	}

	for _, e := range tests {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		if e.addCookie {
			req.AddCookie(e.cookie)
		}
		handler := http.HandlerFunc(app.refreshUsingCookie)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: expected status %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}

func Test_app_deleteRefreshCookie(t *testing.T) {
	req, _ := http.NewRequest("GET", "/logout", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.deleteRefreshCookie)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("expected status %d, but got %d", http.StatusOK, rr.Code)
	}

	foundCookie := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "Host-refresh_token" {
			foundCookie = true
			if c.Expires.After(time.Now()) {
				t.Errorf("expected cookie is deleted, but it was")
			}
		}
	}
	if !foundCookie {
		t.Errorf("Host-refresh_token cookie not found")
	}
}
