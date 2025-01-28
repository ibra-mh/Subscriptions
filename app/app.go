package app

import (
	"database/sql"
	"log"
	"subscriptions/utils"
	"net/http"

	"github.com/gorilla/mux"
)

func InitializeRoute(db *sql.DB) {
	r := mux.NewRouter()

	// Register subscription routes
	SubscriptionRoutes(db, r)
	UserSubscriptionRoutes(db, r)

	// Start the server
	log.Fatal(http.ListenAndServe(":8002", utils.JsonContentTypeMiddleware(r))) // Running on port 8002
}
