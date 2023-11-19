package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func (a *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"POST", "GET", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	mux.Post("/api/payment-intent", a.PaymentIntentHandler)

	mux.Get("/api/get-widget", a.ShowWdigetById)

	// create route for add customer and subscription to add
	mux.Post("/api/add-customer-and-subscription", a.AddCustomerAndSubcsription)

	mux.Post("/api/auth", a.AuthenticationUser)

	mux.Post("/api/check-auth", a.CheckAuth)

	// create route to hgandle reset password requyest
	mux.Post("/api/reset-password", a.ResetPasswordMailer)

	// cretae route to processing reset password
	mux.Post("/api/reset-password-process", a.ProcessResetPassword)

	// create portected route
	mux.Route("/api/admin", func(chi chi.Router) {
		chi.Use(a.AuthUser)

		chi.Post("/virtual-terminal-payment-succeded", a.ProcessSucceddedPayment)

		// create post request to get all sales data
		chi.Post("/showsales", a.GetAllSales)

		// create post requst to get all subscriptions data
		chi.Post("/showsubs", a.GetAllSubscription)

		// create post request to show single sales based on id
		chi.Post("/sales", a.ShowSalesById)

		// create post request for refund purchasing
		chi.Post("/refund", a.RefundPurchasing)

		// create post request for cancel subscription
		chi.Post("/cancel", a.CancleSubscription)

		// create post request to get all users
		chi.Post("/users", a.GetAllUsers)

		// cretae post request to get single user
		chi.Post("/user/{id}", a.GetUser)

		// create post request to update user
		chi.Post("/user/edit/{id}", a.EditUser)

		// create post request to delete user
		chi.Post("/delete/{id}", a.DeleteUser)
	})

	return mux
}
