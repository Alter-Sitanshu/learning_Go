package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/go-chi/chi/v5"
)

func (app *Application) AuthoriseHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	status := http.StatusOK
	ctx := r.Context()
	res := Response{
		Message: "authenticated",
	}

	err := app.store.User().ActivateUser(ctx, token, time.Now())
	if err != nil {
		switch {
		case errors.Is(err, database.ErrTokenExpired):
			res.Message = "Invalid Token/ Expired Token"
			status = http.StatusBadRequest
		default:
			res.Message = "server error"
			status = http.StatusInternalServerError
		}
		log.Printf("server error: %v\n", err.Error())
	}
	jsonResponse(w, status, res)
}
