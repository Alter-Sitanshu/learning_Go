package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application struct {
	config Config
	store  database.Storage
}

type Config struct {
	addr string
	db   DBConfig
}
type DBConfig struct {
	addr         string
	MaxConns     int
	MaxIdleConns int
	MaxIdleTime  int
}

func (app *Application) mount() http.Handler {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(60 * time.Second))

	// Routes for the API
	router.Route("/v1", func(r chi.Router) {

		r.Get("/health", app.HealthCheck)

	})

	return router
}

func (app *Application) run(mux http.Handler) error {
	server := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	fmt.Printf("Server listening on port 8080 at http://localhost:%s\n", app.config.addr)
	return server.ListenAndServe()
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
}
