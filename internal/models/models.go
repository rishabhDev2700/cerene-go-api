package models

import "time"

type User struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Route struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        string    `json:"path"` // LINESTRING stored as WKT
	CreatedAt   time.Time `json:"created_at"`
}

type RecommendedStop struct {
	ID          int64     `json:"id"`
	RouteID     int64     `json:"route_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // viewpoint, shop, landmark, etc.
	Description string    `json:"description"`
	Location    string    `json:"location"` // POINT stored as WKT
	CreatedAt   time.Time `json:"created_at"`
}
