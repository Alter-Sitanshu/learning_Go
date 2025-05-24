package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/go-chi/chi/v5"
)

type PostPayload struct {
	Title   string   `json:"title" validate:"required,max=250"`
	Content string   `json:"content" validate:"required,max=1024"`
	Tags    []string `json:"tags"`
}

type Response struct {
	Message string `json:"message"`
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
}

func (app *Application) CreatPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload PostPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		log.Printf("%s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Message: "Server Error while Parsing",
		})
		return
	}
	if err := Validate.Struct(payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: err.Error(),
		})
		return
	}

	post := database.Post{
		Title:   payload.Title,
		Content: payload.Content,
		// TODO CHANGE AFTER AUTH
		UserID: 1,
		Tags:   payload.Tags,
	}
	ctx := r.Context()
	err := app.store.Post().Create(ctx, &post)
	if err != nil {
		log.Printf("DB Error: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error creating post into database\n: %v", err)))
		return
	}
	if err := WriteJSON(w, post, http.StatusCreated); err != nil {
		log.Printf("Error while encoding post: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error encoding post\n"))
		return
	}
	WriteJSON(w, post, http.StatusCreated)
}

func (app *Application) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		log.Printf("Invalid id param received: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Message: "ID should be integer.",
		})
		return
	}
	ctx := r.Context()
	post, err := app.store.Post().GetPostByID(ctx, id)
	if err != nil {
		log.Printf("DB Error occured: %s", err.Error())
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Response{
			Message: "Post Not Found",
		})
		return
	}
	WriteJSON(w, post, http.StatusOK)
}
