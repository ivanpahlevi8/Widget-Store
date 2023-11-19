package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

// create route function to return http handler
// route can be created uising chi router
func (app *Application) routes() http.Handler {
	// create mux as http handler
	mux := chi.NewRouter()

	// use middleware for session
	mux.Use(app.SessionLoadMiddleware)

	mux.Post("/virtual-terminal-payment-succeded", app.VirtualTerminalSuccededChargePayment)

	// add get request for home
	mux.Get("/", app.Home)

	// add get request for home
	mux.Get("/receipt", app.ShowReceipt)

	mux.Get("/virtual-terminal-receipt", app.VirtualTerminalShowReceipt)

	// add post request
	mux.Post("/payment-succeded", app.SuccededChargePayment)

	// create rout for bronze subscriotin
	mux.Get("/bronze", app.BronzeSubscription)

	// create route for show bronze receipt
	mux.Get("/receipt/bronze", app.ShowBronzeReceipt)

	mux.Get("/login", app.ShowLoginPage)

	mux.Post("/login-process", app.ProcessLogin)

	// create route for logout
	mux.Get("/logout", app.LogoutProcess)

	// crete route to show reset password page
	mux.Get("/reset", app.ResetPassword)

	// create route to show reset password
	mux.Get("/reset-password", app.ShowResetPassword)

	// create route to receiving data from websocket
	mux.Get("/websocket", app.WebsocketEndPoint)

	// create router for admin
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.AuthUser)

		mux.Get("/virtual-terminal", app.VirtualTerminal)

		mux.Get("/all-sales", app.ShowAllSales)

		mux.Get("/all-sub", app.ShowAllSubs)

		mux.Get("/sales/{id}", app.ShowSale)

		mux.Get("/subscription/{id}", app.ShowSubs)

		// create function to show all users
		mux.Get("/all-users", app.ShowAllUsers)

		// create function to show single user
		mux.Get("/user/{id}", app.ShowUser)
	})

	// create route for accessing static file
	httpFile := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", httpFile))

	// create route to show buy one page
	mux.Get("/buy-one", app.OncePage)

	return mux
}
