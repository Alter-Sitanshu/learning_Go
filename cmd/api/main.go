package main

import (
	"log"

	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/Alter-Sitanshu/learning_Go/internal/env"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env", err.Error())
	}

	// Application Configuration
	cfg := Config{
		addr: env.GetString("PORT", ":8080"),
		db: DBConfig{
			addr:         env.GetString("DB_ADDR", "postgres://user:adminpassword@localhost/social?sslmode=disable"),
			MaxConns:     10,
			MaxIdleConns: 5,
			MaxIdleTime:  15, // in minutes
		},
	}

	// Database initialisation
	db, err := database.Mount(
		cfg.db.addr,
		cfg.db.MaxConns,
		cfg.db.MaxIdleConns,
		cfg.db.MaxIdleTime,
	)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()
	psql := database.NewPostgresStorage(db)

	app := &Application{
		config: cfg,
		store:  psql,
	}

	// Server Mux and Routing
	HandlerMux := app.mount()

	// Server initialisation and Configs
	log.Fatal(app.run(HandlerMux))
}
