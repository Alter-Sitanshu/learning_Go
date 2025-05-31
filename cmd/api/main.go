package main

import (
	"log"
	"time"

	"github.com/Alter-Sitanshu/learning_Go/internal/auth"
	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/Alter-Sitanshu/learning_Go/internal/env"
	"github.com/Alter-Sitanshu/learning_Go/internal/mailer"
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
		mail: mailer.SMTPConfig{
			Host:     "smtp.gmail.com",
			Port:     587,
			Username: env.GetString("COMPANY", "example@gmail.com"),
			Password: env.GetString("SMTP_PASS", ""),
			From:     env.GetString("COMP_ADDR", "example@gmail.com"),
			Expiry:   time.Hour * 24 * 3,
		},
		auth: BasicAuthConfig{
			username: env.GetString("ADMIN_USER", "admin"),
			pass:     env.GetString("ADMIN_PASS", "admin"),
			token: TokenConfig{
				secret: env.GetString("APP_SECRET", "example"),
				exp:    time.Hour * 24 * 3,
			},
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
	mailer := mailer.NewSMTPSender(cfg.mail)
	Hostname := "GOSocial"
	jwt := auth.NewAuthenticator(
		cfg.auth.token.secret,
		Hostname,
		Hostname,
	)

	app := &Application{
		config:        cfg,
		store:         psql,
		mailer:        mailer,
		authenticator: jwt,
	}

	// Server Mux and Routing
	HandlerMux := app.mount()

	// Server initialisation and Configs
	log.Fatal(app.run(HandlerMux))
}
