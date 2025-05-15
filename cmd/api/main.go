package main

import (
	"log"

	"github.com/Alter-Sitanshu/learning_Go/internal/env"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env", err.Error())
	}
	cfg := Config{
		addr: env.GetString("PORT", ":8080"),
	}
	app := &Application{
		config: cfg,
	}

	// Server Mux and Routing
	HandlerMux := app.mount()

	// Server initialisation and Configs
	log.Fatal(app.run(HandlerMux))
}
