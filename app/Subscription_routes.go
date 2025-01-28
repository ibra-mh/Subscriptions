package app

import (
	"database/sql"
	"subscriptions/Controllers"
	"github.com/gorilla/mux"
)

func SubscriptionRoutes(db *sql.DB, r *mux.Router) {
	// Subscription Routes
	r.HandleFunc("/subscriptions", controllers.GetSubscriptions(db)).Methods("GET")
	r.HandleFunc("/subscriptions/{id}", controllers.GetSubscriptionByID(db)).Methods("GET")
	r.HandleFunc("/subscriptions", controllers.CreateSubscription(db)).Methods("POST")
	r.HandleFunc("/subscriptions/{id}", controllers.UpdateSubscription(db)).Methods("PUT")
	r.HandleFunc("/subscriptions/{id}", controllers.DeleteSubscription(db)).Methods("DELETE")
}
