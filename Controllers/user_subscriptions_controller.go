package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"subscriptions/models"
	"net/http"
	"github.com/gorilla/mux"
)

// GetUserSubscriptions retrieves all user subscriptions
func GetUserSubscriptions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM user_subscriptions WHERE deleted_at IS NULL")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		userSubscriptions := []models.UserSubscription{}
		for rows.Next() {
			var userSubscription models.UserSubscription
			if err := rows.Scan(&userSubscription.ID, &userSubscription.UserID, &userSubscription.SubscriptionID, &userSubscription.CreatedAt, &userSubscription.UpdatedAt, &userSubscription.DeletedAt); err != nil {
				log.Fatal(err)
			}
			userSubscriptions = append(userSubscriptions, userSubscription)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userSubscriptions)
	}
}

// GetUserSubscriptionByID retrieves a user subscription by ID
func GetUserSubscriptionByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var userSubscription models.UserSubscription
		err := db.QueryRow("SELECT * FROM user_subscriptions WHERE id = $1 AND deleted_at IS NULL", id).
			Scan(&userSubscription.ID, &userSubscription.UserID, &userSubscription.SubscriptionID, &userSubscription.CreatedAt, &userSubscription.UpdatedAt, &userSubscription.DeletedAt)
		if err != nil {
			http.Error(w, "User subscription not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userSubscription)
	}
}

// CreateUserSubscription creates a new user subscription
func CreateUserSubscription(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userSubscription models.UserSubscription
		if err := json.NewDecoder(r.Body).Decode(&userSubscription); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := db.QueryRow("INSERT INTO user_subscriptions (user_id, subscription_id) VALUES ($1, $2) RETURNING id, created_at, updated_at", userSubscription.UserID, userSubscription.SubscriptionID).
			Scan(&userSubscription.ID, &userSubscription.CreatedAt, &userSubscription.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userSubscription)
	}
}

// UpdateUserSubscription updates an existing user subscription
func UpdateUserSubscription(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var userSubscription models.UserSubscription
		if err := json.NewDecoder(r.Body).Decode(&userSubscription); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE user_subscriptions SET user_id = $1, subscription_id = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND deleted_at IS NULL", userSubscription.UserID, userSubscription.SubscriptionID, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userSubscription)
	}
}

// DeleteUserSubscription deletes a user subscription (soft delete)
func DeleteUserSubscription(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE user_subscriptions SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
