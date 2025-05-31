package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Alter-Sitanshu/learning_Go/internal/database"
	"github.com/go-chi/chi/v5"
)

type PostKey string

const postctx PostKey = "post"

type PostPayload struct {
	Title   string   `json:"title" validate:"required,max=250"`
	Content string   `json:"content" validate:"required,max=1024"`
	Tags    []string `json:"tags"`
}

// Payload struct for updating Posts
type PostMutate struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

type Response struct {
	Message string `json:"message"`
}

func (app *Application) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK\n"))
}

func getPostbyctx(r *http.Request) *database.Post {
	post := r.Context().Value(postctx).(*database.Post)
	return post
}

func jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}
	err := WriteJSON(w, status, &envelope{Data: data})
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) CreatPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload PostPayload
	var res Response
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
	user := getUserFromCtx(r)

	post := database.Post{
		Title:   payload.Title,
		Content: payload.Content,
		UserID:  user.ID,
		Tags:    payload.Tags,
	}
	ctx := r.Context()
	err := app.store.Post().Create(ctx, &post)
	if err != nil {
		log.Printf("DB Error: %s", err.Error())
		res.Message = fmt.Sprintf("Error creating post into database: %v", err)
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}
	if err := jsonResponse(w, http.StatusCreated, post); err != nil {
		log.Printf("Error while encoding post: %s", err.Error())
		res.Message = "Error encoding post"
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}
}

func (app *Application) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	res := Response{
		Message: "OK",
	}
	post := getPostbyctx(r)

	comments, err := app.store.Comment().GetComments(r.Context(), post.ID)
	if err != nil {
		log.Printf("DB Error occured: %s", err.Error())
		res.Message = "Error loading comments"
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}
	post.Comments = comments
	jsonResponse(w, http.StatusOK, post)
}

func (app *Application) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	res := Response{
		Message: "Record Deleted",
	}
	post := getPostbyctx(r)

	err := app.store.Post().DeletePost(r.Context(), post.ID)
	if err != nil {
		log.Printf("DB error: %v", err.Error())
		res.Message = "Server error"
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}
	jsonResponse(w, http.StatusOK, res)
}

func (app *Application) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {

	res := Response{
		Message: "Record Updated",
	}
	post := getPostbyctx(r)
	var updatepayload PostMutate

	if err := ReadJSON(w, r, &updatepayload); err != nil {
		res.Message = "Incorrect data format"
		jsonResponse(w, http.StatusBadRequest, res)
		return
	}
	if err := Validate.Struct(updatepayload); err != nil {
		res.Message = "Incorrect data format"
		jsonResponse(w, http.StatusBadRequest, res)
		return
	}

	if updatepayload.Title != nil {
		post.Title = *updatepayload.Title
	}
	if updatepayload.Content != nil {
		post.Content = *updatepayload.Content
	}

	err := app.store.Post().UpdatePost(r.Context(), post)
	if err != nil {
		log.Printf("DB error: %v", err.Error())
		res.Message = "Server Error"
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}

	jsonResponse(w, http.StatusOK, res)
}

func (app *Application) PostMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		postID := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(postID, 10, 64)
		res := Response{
			Message: "OK",
		}
		if err != nil {
			log.Printf("Invalid id param received: %s", err.Error())
			res.Message = "ID should be an integer"
			jsonResponse(w, http.StatusBadRequest, res)
			return
		}
		ctx := r.Context()
		post, err := app.store.Post().GetPostByID(ctx, id)
		if err != nil {
			log.Printf("DB Error occured: %s", err.Error())
			res.Message = "Not Found"
			jsonResponse(w, http.StatusNotFound, res)
			return
		}
		if post.UserID != user.ID {
			log.Printf("Restricted: Request for id: %d, but userid: %d", post.UserID, user.ID)
			res.Message = "Restricted: Request Rejected"
			jsonResponse(w, http.StatusNotAcceptable, res)
			return
		}
		ctx = context.WithValue(ctx, postctx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostbyctx(r)
	var comment database.Comment
	res := Response{
		Message: "Comment added",
	}
	type commentPayload struct {
		Content string `json:"content" validate:"max=100"`
	}
	var payload commentPayload
	user := getUserFromCtx(r)
	userid := user.ID

	if err := ReadJSON(w, r, &payload); err != nil {
		res.Message = "Incorrect data format"
		jsonResponse(w, http.StatusBadRequest, res)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		res.Message = "Validation failed: content may be too long"
		jsonResponse(w, http.StatusBadRequest, res)
		return
	}
	comment.Content = payload.Content
	comment.Postid = post.ID
	comment.Userid = userid

	err := app.store.Comment().CreateComment(r.Context(), &comment)
	if err != nil {
		log.Printf("DB error: %v", err.Error())
		res.Message = "Server Error"
		jsonResponse(w, http.StatusInternalServerError, res)
		return
	}

	jsonResponse(w, http.StatusOK, res)
}
