package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

var Db *Queries

const MigrationFilesURL = "file://pkg/data/migrations"

func GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

}

func InitPostgresDb() {
	sqlc, err := sql.Open("postgres", GetConnectionString())

	if err != nil {
		log.Fatal(err)
	}

	Db = New(sqlc)
}
