package main

import (
	"fmt"
	"go_test_prac/webApp/pkg/data"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

var pathToTemplates = "./templates/"
var uploadPath = "./static/img/"

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var td = make(map[string]any)

	if app.Session.Exists(r.Context(), "test") {
		msg := app.Session.GetString(r.Context(), "test")
		td["test"] = msg
	} else {
		app.Session.Put(r.Context(), "test", "Hit this page at" + time.Now().UTC().String())
	}
	_ = app.render(w, r, "home.page.gohtml", &TemplateData{Data: td})
}

func (app *application) Profile(w http.ResponseWriter, r *http.Request) {
	_ = app.render(w, r, "profile.page.gohtml", &TemplateData{})
}

type TemplateData struct{
	IP string
	Data map[string]any
	Error string
	Flash string
	User data.User
}
func (app *application) render(w http.ResponseWriter, r *http.Request, t string, td *TemplateData) error {
	// parse the template from disk.
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, t), path.Join(pathToTemplates, "base.layout.gohtml"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return err
	}

	td.IP = app.ipFromContext(r.Context())

	td.Error = app.Session.PopString(r.Context(), "error")
	td.Flash = app.Session.PopString(r.Context(), "flash")

	if (app.Session.Exists(r.Context(), "user")) {
		td.User = app.Session.Get(r.Context(), "user").(data.User)
	}

	// execute the template, passing it data if any
	err = parsedTemplate.Execute(w, td)
	if err != nil {
		return err
	}
	return nil
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// validate data
	form := NewForm(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		// redirect toh the login page with error message
		app.Session.Put(r.Context(), "error", "Invalid email or password")// add msg in context
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// get the username and password from the form
	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		// redirect toh the login page with error message
		app.Session.Put(r.Context(), "error", "Invalid Login")// add msg in context
		http.Redirect(w, r, "/", http.StatusSeeOther)// 303, 別ページへの移動
		return
	}

	if !app.authenticate(r, user, password) {
		// redirect toh the login page with error message
		app.Session.Put(r.Context(), "error", "Invalid Login")// add msg in context
		http.Redirect(w, r, "/", http.StatusSeeOther)// 303, 別ページへの移動
		return
	}

	// prevent fixation attack(セッション固定攻撃対策)
	_ = app.Session.RenewToken(r.Context())

	// store success message in session

	// redirect to some other page
	// flashは一時的なメッセージを表示するためのキーフレーズ
	app.Session.Put(r.Context(), "flash", "Successfully logged in!")// add msg in context
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)// 303, 別ページへの移動
}

// DBから取得したユーザのパスワード確認後、セッションにユーザ情報を格納
func (app *application) authenticate(r *http.Request, user *data.User, password string) bool {
	if valid, err := user.PasswordMatches(password); err != nil || !valid {
		return false
	}

	app.Session.Put(r.Context(), "user", user)
	return true
}

func (app *application) UploadProfilePic(w http.ResponseWriter, r *http.Request) {
	files, err := app.UploadFiles(r, uploadPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := app.Session.Get(r.Context(), "user").(data.User)//ミドルウェアに守られているからユーザはnilにならない

	var i = data.UserImage{
		UserID: user.ID,
		FileName: files[0].OriginalFileName,
	}

	_, err = app.DB.InsertUserImage(i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ユーザ情報を更新
	updatadUser, err := app.DB.GetUser(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app.Session.Put(r.Context(), "user", updatadUser)

	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

type UploadedFile struct {
	OriginalFileName string
	FileSize int64
}

// 対象のディレクトリにリクエストから送られてきたファイルを保存する関数
func (app *application) UploadFiles(r *http.Request, uploadDir string) ([]*UploadedFile, error) {
	var uploadedFiles []*UploadedFile

	err := r.ParseMultipartForm(int64(1024 * 1024 * 5)) // 5MB
	if err != nil {
		return nil, fmt.Errorf("the uploaded file is too big. Please choose an image less than 5MB in size")
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func (uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()//　メモリリーク防止

				uploadedFile.OriginalFileName = hdr.Filename

				var outfile *os.File
				defer outfile.Close()//　メモリリーク防止

				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.OriginalFileName)); err != nil {
					return nil, err
				} else {
					fileSize, err := io.Copy(outfile, infile)
					if err != nil {
						return nil, err
					}
					uploadedFile.FileSize = fileSize
				}

				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, err
			}
		}
	}

	return uploadedFiles, nil
}
