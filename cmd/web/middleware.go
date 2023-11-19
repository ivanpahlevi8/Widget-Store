package main

import "net/http"

// create session middlware function
func (app *Application) SessionLoadMiddleware(next http.Handler) http.Handler {
	return sessionManager.LoadAndSave(next)
}

// create middleware function to authenticated user
func (app *Application) AuthUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if user id is exist in session or not
		isExist := app.Session.Exists(r.Context(), "user_id")

		// check user id
		if !isExist {
			// if user id is not exist
			app.infoLog.Println("user not yet login...")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		//if user id exist
		next.ServeHTTP(w, r)
	})
}
