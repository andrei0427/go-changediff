package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/andrei0427/go-changediff/pkg/data"
	mig "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	up := flag.Bool("up", false, "apply migrations")
	down := flag.Bool("down", false, "revert all migrations")
	v := flag.Int("v", 0, "version to force")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	m, err := mig.New(data.MigrationFilesURL, data.GetConnectionString())

	if err != nil {
		log.Fatal(err)
	}

	if *v > 0 {
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
