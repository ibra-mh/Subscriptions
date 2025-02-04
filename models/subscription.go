package models

import "time"


type Subscription struct {
    ID            int       `json:"id"`
    Name          string    `json:"name"`
    ProductID     int       `json:"product_id"`
    LicenseCount  int       `json:"license_count"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    DeletedAt     *time.Time `json:"deleted_at"`
}
