package main

import (
	"database/sql"
	"log"
	"os"
	"subscriptions/app"
	_ "github.com/lib/pq"
)

func main() {
	// Open connection to the PostgreSQL database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create the subscriptions and user_subscriptions tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS subscriptions (
			id SERIAL PRIMARY KEY,
			name VARCHAR NOT NULL,
			product_id INT NOT NULL,
			license_count INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP 
		);

		CREATE TABLE IF NOT EXISTS user_subscriptions (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			subscription_id INT NOT NULL REFERENCES subscriptions(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP 
		);
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize routes and start the server
	app.InitializeRoute(db)
}
