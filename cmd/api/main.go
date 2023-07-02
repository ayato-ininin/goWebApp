package main

import (
	"flag"
	"fmt"
	"go_test_prac/webApp/pkg/repository"
	"go_test_prac/webApp/pkg/repository/dbrepo"
	"log"
	"net/http"
)

const port = 8090

type application struct {
	DSN string
	DB repository.DatabaseRepo
	Domain string
	JWTSecret string
}

func main() {
	var app application
	flag.StringVar(&app.Domain, "domain", "example.com", "Domain for the application, e.g company.com")
	flag.StringVar(&app.DSN, "dsn", "host=localhost user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection")
	flag.StringVar(&app.JWTSecret, "jwt-secret", "2dce505d96a53c5768052ee90fsdf2055657518ad489160df9913f66042e160", "singning secret for JWT")
	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	app.DB = &dbrepo.PostgresDBRepo{DB: conn}

	log.Printf("Starting api on port %d\n", port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
