package dbrepo

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
)

var (
	host = "localhost"
	user = "postgres"
	password = "postgres"
	dbname = "users_test"
	port = "5435"
	dsn = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var resource *dockertest.Resource
var pool *dockertest.Pool
var testDB *sql.DB

func TestMain(m *testing.M) {

	code := m.Run()

	os.Exit(code)
}
