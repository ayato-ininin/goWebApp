package main

import (
	"go_test_prac/webApp/pkg/db"
	"log"
	"os"
	"testing"
)

// testに共通の設定
var app application

func TestMain(m *testing.M) {
	pathToTemplates = "./../../templates/"

	app.Session = getSession() // get a session manager
	app.DSN = "host=localhost user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5"

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app.DB = db.PostgresConn{DB: conn}

	os.Exit(m.Run())
}
