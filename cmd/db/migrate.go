package main

import (
	"flag"
	"fmt"
	"log"

	"database/sql"

	"github.com/andrei0427/go-changediff/internal/data"
	"github.com/joho/godotenv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// SimpleLogger is a custom logger that implements the Logger interface
type SimpleLogger struct{}

// Printf prints the format and args to the standard logger
func (sl *SimpleLogger) Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// Verbose returns true to enable verbose logging
func (sl *SimpleLogger) Verbose() bool {
	return true
}

func main() {
	up := flag.Bool("up", false, "apply migrations")
	down := flag.Bool("down", false, "revert all migrations")
	v := flag.Int("v", -1, "version to force")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("postgres", data.GetConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(data.MigrationFilesURL, "postgres", driver)
	if err != nil {
		log.Fatal(err)
	}

	m.Log = &SimpleLogger{}
	defer m.Close()

	if *v > -1 {
		if err := m.Force(*v); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Migrations currently versioned at %v", *v)
	} else if *up {
		if err := m.Up(); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Migrations successfully applied")
	} else if *down {
		if err := m.Down(); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Migrations successfully reverted")
	} else {
		log.Fatal("Invalid or missing operation")
	}
}
