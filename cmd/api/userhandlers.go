package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/Alter-Sitanshu/learning_Go/internal/mailer"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserKey string

const userctx UserKey = "user"

type UserPayload struct {
	Name     string `json:"name" validate:"min=1,max=255"`
	Password string `json:"password" validate:"min=8,max=72"`
	Age      int    `json:"age" validate:"min=1,max=100"`
	Email    string `json:"email" validate:"min=12"`
}

func getUserFromCtx(r *http.Request) *database.User {
	user, _ := r.Context().Value(userctx).(*database.User)
	return user
}

func (app *Application) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userid := chi.URLParam(r, "userID")
	id, err := strconv.ParseInt(userid, 10, 64)
	res := Response{}
	if err != nil {
		log.Println("Could not parse user id.")
		res.Message = "id should only contain integers."
		jsonResponse(w, http.StatusBadRequest, res)
		return
	}

	user, err := app.store.User().GetUserByID(ctx, id)
	if err != nil {
		log.Printf("DB error: %v\n", err.Error())
		switch {
		case errors.Is(err, database.ErrNotFound):
			res.Message = "User not found"
			jsonResponse(w, http.StatusNotFound, res)
			return
		default:
			res.Message = "Server error"
			jsonResponse(w, http.StatusInternalServerError, res)
			return
		}
	}
	jsonResponse(w, http.StatusOK, user)
}

func (app *Application) FollowUser(w http.ResponseWriter, r *http.Request) {
	targetID, _ := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	user := getUserFromCtx(r)

	res := Response{
		Message: fmt.Sprintf("Followed user with id: %d", targetID),
	}

	ctx := r.Context()
	err := app.store.User().Follow(ctx, targetID, user.ID)
	if err != nil {
		log.Printf("Error: %v\n", err.Error())
		res.Message = err.Error()
		jsonResponse(w, http.StatusBadRequest, res)
	}

	jsonResponse(w, http.StatusNoContent, res)
}

func (app *Application) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	targetID, _ := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	user := getUserFromCtx(r)

	res := Response{
		Message: fmt.Sprintf("Unfollowed user with id: %d", targetID),
	}

	ctx := r.Context()
	err := app.store.User().Unfollow(ctx, targetID, user.ID)
	if err != nil {
		log.Printf("Error: %v\n", err.Error())
		res.Message = err.Error()
		jsonResponse(w, http.StatusBadRequest, res)
	}

	jsonResponse(w, http.StatusNoContent, res)
}

func (app *Application) GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	userid := getUserFromCtx(r).ID
	fq := &database.FilteringQuery{}
	res := Response{
		Message: "Feed fetched",
	}
	err := fq.Parse(r)
	if err != nil {
		log.Printf("Bad Request: %v\n", err.Error())
		res.Message = "Bad request: Error while parsing"
		jsonResponse(w, http.StatusBadRequest, res)
		return
	}
	if err = Validate.Struct(fq); err != nil {
		log.Printf("Bad Request: %v\n", err.Error())
		res.Message = err.Error()
		jsonResponse(w, http.StatusBadRequest, res)
		return
	}

	feed, err := app.store.User().GetFeed(r.Context(), userid, fq)

	if err != nil {
		log.Printf("Server Error: %v\n", err.Error())
		res.Message = "server error"
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}

	jsonResponse(w, http.StatusOK, feed)
}

func (app *Application) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload UserPayload
	ctx := r.Context()
	res := Response{
		Message: "user created",
	}
	status := http.StatusCreated
	err := ReadJSON(w, r, &payload)
	if err != nil {
		log.Printf("bad request: %v\n", err.Error())
		res.Message = "bad request: Error while parsing"
		status = http.StatusBadRequest
		jsonResponse(w, status, res)
		return
	}
	if err = Validate.Struct(payload); err != nil {
		log.Printf("bad request: %v\n", err.Error())
		res.Message = err.Error()
		status = http.StatusBadRequest
		jsonResponse(w, status, res)
		return
	}

	user := &database.User{
		Name:  payload.Name,
		Email: payload.Email,
		Age:   payload.Age,
		Role:  1,
	}

	if err = user.Password.Hash(payload.Password); err != nil {
		log.Printf("server error: %v\n", err.Error())
		res.Message = "internal server error"
		status = http.StatusInternalServerError
		jsonResponse(w, status, res)
		return
	}

	plainToken := uuid.New().String()
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	err = app.store.User().CreateAndInvite(ctx, user, hashToken, app.config.mail.Expiry)
	if err != nil {
		app.store.User().DeleteUser(ctx, user)
		switch err {
		case database.ErrDupliMail:
			res.Message = "email already exists"
			status = http.StatusBadRequest
		case database.ErrDupliName:
			res.Message = "name taken"
			status = http.StatusBadRequest
		default:
			res.Message = "internal server error"
			status = http.StatusInternalServerError
		}
		log.Printf("error occured while creating user: %v\n", err.Error())
		jsonResponse(w, status, res)
		return
	}

	// res.Message = plainToken
	// Can send the code through a JSON response to verrify
	// if err = jsonResponse(w, http.StatusCreated, res); err != nil {
	// 	log.Printf("server error: %v\n", err.Error())
	// 	res.Message = "internal server error"
	// 	jsonResponse(w, http.StatusInternalServerError, res)
	// 	return
	// }

	req := mailer.EmailRequest{
		To:      user.Email,
		Subject: "Auth-Bearer/Token",
		Body:    fmt.Sprintf("Your user verification token is: %v\nExpires in: 3 days", plainToken),
	}
	err = app.mailer.SendEmail(req)
	if err != nil {
		app.store.User().DeleteUser(ctx, user) // (SAGA) pattern
		log.Printf("encountered error sending mail: %v\n", err.Error())
		res.Message = "error sending email, retry"
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}

}

func (app *Application) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromCtx(r)
	err := app.store.User().DeleteUser(r.Context(), user)
	res := Response{}
	if err != nil {
		log.Printf("server error: %v\n", err.Error())
		res.Message = "internal server error"
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}

	res.Message = "User Deleted"
	jsonResponse(w, http.StatusNoContent, res)
}
