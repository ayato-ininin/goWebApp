package main

import (
	"go_test_prac/webApp/pkg/repository/dbrepo"
	"os"
	"testing"
)

// testに共通の設定
var app application

func TestMain(m *testing.M) {
	pathToTemplates = "./../../templates/"

	app.Session = getSession() // get a session manager
	app.DB = &dbrepo.TestDBRepo{}

	os.Exit(m.Run())
}
