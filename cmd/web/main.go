package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	DSN string // data source name(PW等含む)
	DB *sql.DB // DBconnetion
	Session *scs.SessionManager
}
func main() {
	// set up an app config
	app := application{}

	// flagパッケージを使って、コマンドライン引数をパース
	// go run ./cmd/web -dsn="user=yourusername password=yourpassword dbname=yourdbname sslmode=disable"
	// 上記指定しなければ、下記デフォルト値が使用される
	flag.StringVar(&app.DSN, "dsn", "host=localhost user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection")
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	app.DB = conn

	// get a session manager
	app.Session = getSession()

	// print out a message
	log.Println("starting server on :8080...")

	// start the server
	err = http.ListenAndServe(":8080", app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
