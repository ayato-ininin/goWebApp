package main

import (
	"go_test_prac/webApp/pkg/data"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
)

var pathToTemplates = "./templates/"

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

	log.Println(password, user.FirstName)

	// authenticate the user
	// if not authenticated redirect to login page with error message

	// prevent fixation attack(セッション固定攻撃対策)
	_ = app.Session.RenewToken(r.Context())

	// store success message in session

	// redirect to some other page
	// flashは一時的なメッセージを表示するためのキーフレーズ
	app.Session.Put(r.Context(), "flash", "Successfully logged in!")// add msg in context
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)// 303, 別ページへの移動
}
