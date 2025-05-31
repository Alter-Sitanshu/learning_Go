package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alter-Sitanshu/learning_Go/internal/auth"
	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/Alter-Sitanshu/learning_Go/internal/mailer"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application struct {
	config        Config
	store         database.Storage
	mailer        *mailer.SMTPSender
	authenticator *auth.Authenticator
}

type Config struct {
	addr string
	db   DBConfig
	mail mailer.SMTPConfig
	auth BasicAuthConfig
}

type BasicAuthConfig struct {
	username string
	pass     string
	token    TokenConfig
}

type TokenConfig struct {
	secret string
	exp    time.Duration
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
		r.With(app.BasicAuthMiddleware()).Get("/health", app.HealthCheck)

		r.Route("/post", func(r chi.Router) {
			r.Use(app.AuthorizationMiddleware)
			r.Post("/", app.CreatPostHandler)
			r.Route("/{id}", func(r chi.Router) {
				// MIDDLEWARE TO ACCESS THE ID AND FETCH POST
				r.Use(app.PostMiddleware)

				r.Get("/", app.GetPostHandler)
				r.Delete("/", app.checkRoleMiddleware("admin", app.DeletePostHandler))
				r.Patch("/", app.checkRoleMiddleware("moderator", app.UpdatePostHandler))
				r.Post("/comment", app.CreateCommentHandler)
			})

		})
		r.Route("/users", func(r chi.Router) {
			// User Middleware
			r.Use(app.AuthorizationMiddleware)
			r.Route("/{userID}", func(r chi.Router) {

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
			r.Post("/token", app.JWTHandler)
			r.Put("/activate/{token}", app.ActivateUserHandler)
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

	// graceful shutdown
	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit // indefinitely waits for quit to end a signal
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
		defer cancel()

		log.Printf("signal caught: %s", s.String())
		shutdown <- server.Shutdown(ctx)
	}()

	fmt.Printf("Server listening at http://localhost%s\n", app.config.addr)
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		log.Println(err.Error())
	}

	log.Println("Server Shutdown successful")
	return nil
}
