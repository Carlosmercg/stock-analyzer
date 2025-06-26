package database

import (
	"database/sql"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var DB *bun.DB

func InitCockroach() *bun.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error cargando .env: %v", err)
	}

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")     // sin "postgresql://"
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	encodedPass := url.QueryEscape(pass)

	dsn := "postgresql://" + user + ":" + encodedPass + "@" + host + ":" + port + "/" + name + "?sslmode=verify-full"

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	DB = bun.NewDB(sqldb, pgdialect.New())

	log.Println("✅ Conectado a CockroachDB con éxito")

	return DB
}
