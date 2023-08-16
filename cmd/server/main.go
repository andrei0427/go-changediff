package main

import (
	"log"

	"github.com/andrei0427/go-changediff/internal/app"
	"github.com/andrei0427/go-changediff/internal/app/handlers"
)

func main() {
	app := app.NewApp()

	handlers.InitRoutes(app)

	if err := app.Fiber.Listen("localhost:8000"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
