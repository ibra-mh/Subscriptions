package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"subscriptions/models"
	"net/http"
	"github.com/gorilla/mux"
)

// GetSubscriptions retrieves all subscriptions
func GetSubscriptions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM subscriptions WHERE deleted_at IS NULL")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		subscriptions := []models.Subscription{}
		for rows.Next() {
			var subscription models.Subscription
			if err := rows.Scan(&subscription.ID, &subscription.Name, &subscription.ProductID, &subscription.LicenseCount, &subscription.CreatedAt, &subscription.UpdatedAt, &subscription.DeletedAt); err != nil {
				log.Fatal(err)
			}
			subscriptions = append(subscriptions, subscription)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subscriptions)
	}
}

// GetSubscriptionByID retrieves a subscription by ID
func GetSubscriptionByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var subscription models.Subscription
		err := db.QueryRow("SELECT * FROM subscriptions WHERE id = $1 AND deleted_at IS NULL", id).
			Scan(&subscription.ID, &subscription.Name, &subscription.ProductID, &subscription.LicenseCount, &subscription.CreatedAt, &subscription.UpdatedAt, &subscription.DeletedAt)
		if err != nil {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subscription)
	}
}

// CreateSubscription creates a new subscription
func CreateSubscription(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var subscription models.Subscription
		if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := db.QueryRow("INSERT INTO subscriptions (name, product_id, license_count) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at", subscription.Name, subscription.ProductID, subscription.LicenseCount).
			Scan(&subscription.ID, &subscription.CreatedAt, &subscription.UpdatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subscription)
	}
}

// UpdateSubscription updates an existing subscription
func UpdateSubscription(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var subscription models.Subscription
		if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("UPDATE subscriptions SET name = $1, product_id = $2, license_count = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4 AND deleted_at IS NULL", subscription.Name, subscription.ProductID, subscription.LicenseCount, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(subscription)
	}
}

// DeleteSubscription deletes a subscription (soft delete)
func DeleteSubscription(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE subscriptions SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
