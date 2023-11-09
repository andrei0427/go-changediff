package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
)

const MigrationFilesURL = "file://internal/data/migrations"

func GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

}

func InitPostgresDb() *sql.DB {
	sqlc, err := sql.Open("postgres", GetConnectionString())

	if err != nil {
		log.Fatal(err)
	}

	return sqlc
}
