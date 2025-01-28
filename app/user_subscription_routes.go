package app

import (
	"database/sql"
	"subscriptions/Controllers"
	"github.com/gorilla/mux"
)

func UserSubscriptionRoutes(db *sql.DB, r *mux.Router) {
	// User Subscription Routes
	r.HandleFunc("/user_subscriptions", controllers.GetUserSubscriptions(db)).Methods("GET")
	r.HandleFunc("/user_subscriptions/{id}", controllers.GetUserSubscriptionByID(db)).Methods("GET")
	r.HandleFunc("/user_subscriptions", controllers.CreateUserSubscription(db)).Methods("POST")
	r.HandleFunc("/user_subscriptions/{id}", controllers.UpdateUserSubscription(db)).Methods("PUT")
	r.HandleFunc("/user_subscriptions/{id}", controllers.DeleteUserSubscription(db)).Methods("DELETE")
}
