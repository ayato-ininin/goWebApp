package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	var theTests = []struct{
		name string
		url string
		expectedStatusCode int
		expectedURL string
		expectedFirstStatusCode int
	}{
		{"home", "/", http.StatusOK, "/", http.StatusOK},
		{"404", "/fish", http.StatusNotFound, "/fish", http.StatusNotFound},
		{"profile", "/user/profile", http.StatusOK, "/", http.StatusTemporaryRedirect},
	}

	routes := app.routes()

	// create a test server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// デフォルトのts.clientはデフォルトでリダイレクトを追跡し、Locationヘッダをリダイレクト先に書き換える。
	// なのでリダイレクト前のレスポンスを取得するためには、リダイレクトを追跡しないclientを作成する必要がある。
	// checkRedirectでリダイレクトの動作をカスタマイズしている。e.expectedFirstStatusCode用。
	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // don't follow redirects
		},
	}

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

		if resp.Request.URL.Path != e.expectedURL {
			t.Errorf("%s: expected final url of %s but got %s", e.name, e.expectedURL, resp.Request.URL.Path)
		}

		resp2, _ := client.Get(ts.URL + e.url)
		if resp2.StatusCode != e.expectedFirstStatusCode {
			t.Errorf("%s: expected first returned code to be %d but got %d", e.name, e.expectedFirstStatusCode, resp2.StatusCode)
		}
	}
}

func TestAppHome(t *testing.T) {
	var tests = []struct{
		name string
		putInSession string
		expectedHTML string
	}{
		{"first visit", "", `<small>From Session:`},
		{"second visit", "hello, world", `<small>From Session: hello, world`},
	}

	for _, e := range tests {
		// create a request
		req, _ := http.NewRequest("GET", "/", nil)
		req = addContextAndSessionToRequest(req, app) // middlewareのかわりに追加設定
		_ = app.Session.Destroy(req.Context())// sessionをクリア

		// 二回目以降を想定してセット
		if e.putInSession != "" {
			app.Session.Put(req.Context(), "test", e.putInSession)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.Home)
		handler.ServeHTTP(rr, req)

		// check status code
		if rr.Code != http.StatusOK {
			t.Errorf("Home returned wrong status code: got %v, want %v", rr.Code, http.StatusOK)
		}

		// 正しいhtmlが返ってきているか確認
		body, _ := io.ReadAll(rr.Body)
		if !strings.Contains(string(body), e.expectedHTML) {
			t.Errorf("%s: did not find %s in response body", e.name, e.expectedHTML)
		}
	}
}


// ./tempaltes/ファイルにないテンプレートを指定した場合のテスト
func TestApp_renderWithBadTemplate(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req = addContextAndSessionToRequest(req, app) // middlewareのかわりに追加設定
	rr := httptest.NewRecorder()

	err := app.render(rr, req, "bad.page.gohtml", &TemplateData{})
	if err == nil {
		t.Error("expected error from bad template, but did not get one")
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

func Test_app_Login(t *testing.T) {
	tests := []struct {
		name string
		postedData url.Values
		expectedStatusCode int
		expectedLocation string
	}{
		{
			name: "valid login",
			postedData: url.Values{
				"email": {"admin@example.com"},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation: "/user/profile",
		},
		{
			name: "missing form data",
			postedData: url.Values{
				"email": {""},
				"password": {""},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation: "/",
		},
		{
			name: "user not found",
			postedData: url.Values{
				"email": {"you@gmail.com"},
				"password": {"password"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation: "/",
		},
		{
			name: "bad credentials",
			postedData: url.Values{
				"email": {"admin@example.com"},
				"password": {"badpw"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLocation: "/",
		},
	}

	for _, e := range tests {
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(e.postedData.Encode()))
		req = addContextAndSessionToRequest(req, app) // middlewareのかわりに追加設定
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.Login)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: returned wrong status code; expected %d, got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		actualLocation, err := rr.Result().Location()
		if err == nil {
			if actualLocation.String() != e.expectedLocation {
				t.Errorf("%s: returned wrong location; expected %s, got %s", e.name, e.expectedLocation, actualLocation.String())
			}
		} else {
			t.Errorf("%s: no location header set", e.name)
		}
	}
}


func Test_app_UploadFiles(t *testing.T) {
	// set up pipe
	pr, pw := io.Pipe()//writerに書き込んだデータをreaderで読み込めるようにする(バッファ)

	// create a new writer, of type *io.Writer
	writer := multipart.NewWriter(pw) // multipart/form-dataでリクエストを送るのをシミュレートするため。formでテンプレートから送信されるのと同じ形式にする必要がある。

	// create a waitgroup, and add 1 to it
	wg := &sync.WaitGroup{}
	wg.Add(1)

	// simulate uploading a file using a goroutine and our writer
	go simulatePNGUpload("./testdata/img.png", writer, t, wg)//goroutineを使うのは同時に複数のファイルをアップロードすることを想定しているため

	// read from the pipe which recieves the file
	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	// call app.UploadFiles()、この上まではリクエストの生成を再現し、それをapp.UploadFiles()に渡している
	UploadedFiles, err := app.UploadFiles(request, "./testdata/uploads/")
	if err != nil {
		t.Error(err)
	}

	// perform our tests
	// この関数の評価としては、来たリクエストのファイルが、ちゃんと意図したディレクトリに保存されているか。
	if _, err := os.Stat(fmt.Sprintf("./testdata/uploads/%s", UploadedFiles[0].OriginalFileName)); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s",err.Error())
	}

	// clean up
	_ = os.Remove(fmt.Sprintf("./testdata/uploads/%s", UploadedFiles[0].OriginalFileName))
}

// .testData/img.pngを予め用意しておき、それをmultipart.Writerに書き込む
// ここでやっているのは、multipart/form-dataで<form>タグから送信されるリクエストを再現するためのもの
// img.pngがform-dataのfileフィールドにあたる
func simulatePNGUpload(fileToUpload string, writer *multipart.Writer, t *testing.T, wg *sync.WaitGroup) {
	defer writer.Close()
	defer wg.Done()

	// create the form data filed 'file' with value being filename
	part, err := writer.CreateFormFile("file", path.Base(fileToUpload))
	if err != nil {
		t.Error(err)
	}

	// open the file to be uploaded
	f, err := os.Open(fileToUpload)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	// decode the image
	img, _, err := image.Decode(f)
	if err != nil {
		t.Error("error decoding image", err)
	}

	// write the png to our io.Writer
	err = png.Encode(part, img)
}

func Test_app_UploadProfilePic(t *testing.T) {
	uploadPath = "./testdata/uploads/"
}
