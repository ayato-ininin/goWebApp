package main

import (
	"os"
	"testing"
)

// testに共通の設定
var app application

func TestMain(m *testing.M) {
	app.Session = getSession() // get a session manager

	os.Exit(m.Run())
}
