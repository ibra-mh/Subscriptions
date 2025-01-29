package models

import "time"

type UserSubscription struct {
	ID               int        `json:"id"`
	UserID           int        `json:"user_id"`
	SubscriptionID   int        `json:"subscription_id"`
	SubscriptionName string     `json:"name"`
	ProductID        int        `json:"product_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at"`
}
