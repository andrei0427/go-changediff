package main

import (
	"log"

	"github.com/andrei0427/go-changediff/internal/app"
	"github.com/andrei0427/go-changediff/internal/app/handlers"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	app := app.NewApp()

	handlers.InitRoutes(app)

	if err := app.Fiber.Listen("localhost:8000"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
