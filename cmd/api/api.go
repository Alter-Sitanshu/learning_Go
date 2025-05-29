package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/Alter-Sitanshu/learning_Go/internal/mailer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application struct {
	config Config
	store  database.Storage
	mailer *mailer.SMTPSender
}

type Config struct {
	addr string
	db   DBConfig
	mail mailer.SMTPConfig
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
		r.Route("/post", func(r chi.Router) {
			r.Post("/", app.CreatPostHandler)
			r.Route("/{id}", func(r chi.Router) {
				// MIDDLEWARE TO ACCESS THE ID AND FETCH POST
				r.Use(app.PostMiddleware)

				r.Get("/", app.GetPostHandler)
				r.Delete("/", app.DeletePostHandler)
				r.Patch("/", app.UpdatePostHandler)
				r.Post("/comment", app.CreateCommentHandler)
			})

		})
		r.Route("/users", func(r chi.Router) {
			r.Route("/{userID}", func(r chi.Router) {
				// User Middleware
				r.Use(app.UserctxMiddleware)

				r.Get("/", app.GetUserHandler)
				r.Put("/follow", app.FollowUser)
				r.Put("/unfollow", app.UnfollowUser)
				r.Delete("/", app.DeleteUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.GetFeedHandler)
			})
		})
		r.Route("/auth", func(r chi.Router) {
			r.Put("/activate/{token}", app.AuthoriseHandler)
			r.Post("/user", app.CreateUserHandler)
		})

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

	fmt.Printf("Server listening at http://localhost%s\n", app.config.addr)
	return server.ListenAndServe()
}
