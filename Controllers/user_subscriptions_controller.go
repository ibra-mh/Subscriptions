package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/mux"
	"net/http"
	"subscriptions/models"
)

// GetUserSubscriptions retrieves all user subscriptions

func GetUserSubscriptions(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")   // Get user_id from query parameter
		fmt.Println("UserID parameter:", userID) // Debugging log

		query := `
            SELECT 
                us.id, us.user_id, us.subscription_id, us.created_at, us.updated_at, us.deleted_at, 
                s.name, s.product_id
            FROM 
                user_subscriptions us
            JOIN 
                subscriptions s ON us.subscription_id = s.id
            WHERE 
                us.deleted_at IS NULL`

		var rows *sql.Rows
		var err error

		// If user_id is provided, filter by user_id
		if userID != "" {
			query += " AND us.user_id = $1" // Add condition for user_id
			rows, err = db.Query(query, userID)
		} else {
			rows, err = db.Query(query) // Fetch all if no user_id is provided
		}

		if err != nil {
			log.Println("Database query error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var userSubscriptions []models.UserSubscription
		for rows.Next() {
			var userSubscription models.UserSubscription
			if err := rows.Scan(
				&userSubscription.ID, &userSubscription.UserID, &userSubscription.SubscriptionID,
				&userSubscription.CreatedAt, &userSubscription.UpdatedAt, &userSubscription.DeletedAt,
				&userSubscription.SubscriptionName, &userSubscription.ProductID); err != nil {
				log.Println("Error scanning row:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			userSubscriptions = append(userSubscriptions, userSubscription)
		}

		if err := rows.Err(); err != nil {
			log.Println("Rows iteration error:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userSubscriptions) // Send the result as JSON
	}
}

// GetUserSubscriptionByID retrieves a user subscription by ID

func GetUserSubscriptionByID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var userSubscription models.UserSubscription
		var name string
		var productID int

		err := db.QueryRow(`
			SELECT 
				us.id, us.user_id, us.subscription_id, us.created_at, us.updated_at, us.deleted_at, 
				s.name, s.product_id
			FROM 
				user_subscriptions us
			JOIN 
				subscriptions s ON us.subscription_id = s.id
			WHERE 
				us.id = $1 AND us.deleted_at IS NULL
		`, id).
			Scan(&userSubscription.ID, &userSubscription.UserID, &userSubscription.SubscriptionID, &userSubscription.CreatedAt, &userSubscription.UpdatedAt, &userSubscription.DeletedAt, &name, &productID)
		if err != nil {
			http.Error(w, "User subscription not found", http.StatusNotFound)
			return
		}

		// Adding name and product_id to the response
		userSubscription.SubscriptionName = name
		userSubscription.ProductID = productID

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userSubscription)
	}
}

// CreateUserSubscription creates a new user subscription

func CreateUserSubscription(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var userSubscription models.UserSubscription
		if err := json.NewDecoder(r.Body).Decode(&userSubscription); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Check if the subscription exists and get its license_count
		var licenseCount int
		err := db.QueryRow("SELECT license_count FROM subscriptions WHERE id = $1 AND deleted_at IS NULL", userSubscription.SubscriptionID).Scan(&licenseCount)
		if err != nil {
			http.Error(w, "Subscription not found", http.StatusNotFound)
			return
		}

		// Count how many user subscriptions currently exist for this subscription
		var currentSubscriptions int
		err = db.QueryRow("SELECT COUNT(*) FROM user_subscriptions WHERE subscription_id = $1 AND deleted_at IS NULL", userSubscription.SubscriptionID).Scan(&currentSubscriptions)
		if err != nil {
			http.Error(w, "Failed to check current subscriptions", http.StatusInternalServerError)
			return
		}

		// Check if licenses are available
		if currentSubscriptions >= licenseCount {
			http.Error(w, "No licenses available for this subscription", http.StatusForbidden)
			return
		}

		// Insert the new user subscription
		err = db.QueryRow(
			"INSERT INTO user_subscriptions (user_id, subscription_id, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id",
			userSubscription.UserID, userSubscription.SubscriptionID,
		).Scan(&userSubscription.ID)
		if err != nil {
			http.Error(w, "Failed to create user subscription", http.StatusInternalServerError)
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
