package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type JWTpayload struct {
	Email    string `json:"email" validate:"min=12"`
	Password string `json:"pass" validate:"min=8,max=72"`
}

func (app *Application) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	status := http.StatusOK
	ctx := r.Context()
	res := Response{
		Message: "user activated",
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

func (app *Application) JWTHandler(w http.ResponseWriter, r *http.Request) {
	// parse the payload and get the user from it
	// encode the creds into the token with claims
	// add expiry to the token
	var payload JWTpayload
	var res Response
	if err := ReadJSON(w, r, &payload); err != nil {
		log.Printf("Error: %v\n", err.Error())
		res.Message = err.Error()
		jsonResponse(w, http.StatusBadRequest, res)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		log.Printf("Error: %v\n", err.Error())
		res.Message = err.Error()
		jsonResponse(w, http.StatusBadRequest, res)
		return
	}

	user, err := app.store.User().GetUserByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case database.ErrNotFound:
			res.Message = "invalid credentials"
			jsonResponse(w, http.StatusUnauthorized, res)
		default:
			res.Message = "server error"
			jsonResponse(w, http.StatusInternalServerError, res)
		}
		return
	}
	err = user.Password.CheckPassword(payload.Password)
	if err != nil {
		res.Message = "invalid credentials"
		jsonResponse(w, http.StatusUnauthorized, res)
		return
	}
	claims := jwt.MapClaims{
		"sub": user.ID,
		"iss": "GOSocial",
		"aud": "GOSocial",
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"nbf": time.Now().Unix(),
		"iat": time.Now().Unix(),
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		res.Message = "server error"
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}
	res.Message = fmt.Sprintf("Token Created: %s", token)
	jsonResponse(w, http.StatusCreated, res)

}

func (app *Application) checkRole(ctx context.Context, RequiredRole string, role int) (bool, error) {
	fetch_role, err := app.store.Role().GetRole(ctx, RequiredRole)
	if err != nil {
		return false, err
	}
	return role >= fetch_role.Level, nil
}
