package main

import (
	"log"
	"net/http"
)

// create middleware function to authenticated user
func (app *application) AuthUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// auth user
		user, err := app.ValidateToken(r)

		// check for an error
		if err != nil {
			log.Println("error when authenticated user with token invalid : ", err)
			if user.Email == "" {
				user.Email = "error@error.com"
			}
			app.InvalidAuthUser(w, user.Email)
			return
		}

		//if valid
		next.ServeHTTP(w, r)
	})
}
