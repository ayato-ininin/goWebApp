package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	var theTests = []struct{
		name string
		url string
		expectedStatusCode int
	}{
		{"home", "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound},
	}

	routes := app.routes()

	// create a test server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	pathToTemplates = "./../../templates/"

	// range through the tests
	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}

		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestAppHome(t *testing.T) {
	// create a request
	req, _ := http.NewRequest("GET", "/", nil)
	req = addContextAndSessionToRequest(req, app) // middlewareのかわりに追加設定

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.Home)
	handler.ServeHTTP(rr, req)

	// check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Home returned wrong status code: got %v, want %v", rr.Code, http.StatusOK)
	}

	// 正しいhtmlが返ってきているか確認
	body, _ := io.ReadAll(rr.Body)
	if !strings.Contains(string(body), `<small>From Session:`) {
		t.Error("did not find correct text in html")
	}
}

func getCtx(req *http.Request) context.Context {
	ctx := context.WithValue(req.Context(), contextUserKey, "unknown")
	return ctx
}

// handlerに渡すので、テストリクエストにmiddlewareで追加しているものを設定しないとエラーになる。
func addContextAndSessionToRequest(req *http.Request, app application) *http.Request {
	// req = req.WithContext(req.Context()) →getCtx(req)に変更しないと、ipFromContextでnil参照してpanic
	req = req.WithContext(getCtx(req)) // contextデータをテストリクエストに追加
	ctx, _ := app.Session.Load(req.Context(), req.Header.Get("X-Session")) // sessionデータをテストリクエストに追加
	return req.WithContext(ctx)
}
