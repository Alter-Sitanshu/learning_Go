package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *Application) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var res Response
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			res.Message = "Auth Header absent"
			jsonResponse(w, http.StatusUnauthorized, res)
			return
		}

		CredSplit := strings.Split(authHeader, " ")
		if len(CredSplit) != 2 || CredSplit[0] != "Bearer" {
			res.Message = "Auth Header malformed"
			jsonResponse(w, http.StatusUnauthorized, res)
			return
		}
		token := CredSplit[1]
		JWTtoken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			res.Message = fmt.Sprintf("unabale to parse token: %v", err.Error())
			jsonResponse(w, http.StatusUnauthorized, res)
			return
		}
		TokenSub, _ := JWTtoken.Claims.(jwt.MapClaims).GetSubject()
		userID, err := strconv.ParseInt(TokenSub, 10, 64)
		if err != nil {
			res.Message = "unauthorised"
			jsonResponse(w, http.StatusUnauthorized, res)
			return
		}
		user, err := app.store.User().GetUserByID(ctx, userID)

		if err != nil {
			res.Message = "unauthorised"
			jsonResponse(w, http.StatusUnauthorized, res)
			return
		}

		ctx = context.WithValue(ctx, userctx, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *Application) checkRoleMiddleware(RequiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromCtx(r)
		post := getPostbyctx(r)
		var res Response
		if post.UserID != user.ID {
			res.Message = "restricted"
			jsonResponse(w, http.StatusForbidden, res)
			return
		}

		allowed, err := app.checkRole(r.Context(), RequiredRole, user.Role)
		if err != nil {
			res.Message = "server error"
			jsonResponse(w, http.StatusInternalServerError, res)
			return
		}

		if !allowed {
			res.Message = "restricted"
			jsonResponse(w, http.StatusForbidden, res)
			return
		}

		next.ServeHTTP(w, r)

	})
}

func (app *Application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				jsonResponse(w, http.StatusUnauthorized, "Auth Header absent")
				return
			}
			// parse it
			authdata := strings.Split(authHeader, " ")
			if len(authdata) != 2 || authdata[0] != "Basic" {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				jsonResponse(w, http.StatusUnauthorized, "Auth header malformed")
				return
			}
			// decode it and check the creds
			decode, err := base64.StdEncoding.DecodeString(authdata[1])
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				jsonResponse(w, http.StatusUnauthorized, err)
				return
			}
			creds := strings.SplitN(string(decode), ":", 2)
			username := app.config.auth.username
			pass := app.config.auth.pass

			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				jsonResponse(w, http.StatusUnauthorized, "Not Authorized")
				return
			}

			// serve the route
			next.ServeHTTP(w, r)
		})
	}
}
