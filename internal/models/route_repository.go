package models

import (
	"context"
	"database/sql"
)

type RouteRepository interface {
	Create(ctx context.Context, Route *Route) error
	GetByID(ctx context.Context, id int) (*Route, error)
	GetNearby(ctx context.Context, lat, lng, distance float64) (*[]Route, error)
	Update(ctx context.Context, Card *Route) error
	Delete(ctx context.Context, id int) error
}

type routeRepository struct {
	db *sql.DB
}

func NewRouteRepository(db *sql.DB) RouteRepository {
	return &routeRepository{db: db}
}

// Create inserts a new route into the database
func (r *routeRepository) Create(ctx context.Context, route *Route) error {
	query := `INSERT INTO routes (id, name, description, path, created_at) 
              VALUES ($1, $2, $3, ST_GeomFromText($4, 4326), NOW())`
	_, err := r.db.ExecContext(ctx, query, route.ID, route.Name, route.Description, route.Path)
	return err
}

// GetByID fetches a route by its ID and ensures it belongs to the user
func (r *routeRepository) GetByID(ctx context.Context, id int) (*Route, error) {
	query := `SELECT id, name, description, ST_AsText(path), created_at 
              FROM routes WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var route Route
	err := row.Scan(&route.ID, &route.Name, &route.Description, &route.Path, &route.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &route, nil
}

// GetNearby retrieves routes close to a given route (simple example using ST_DWithin)
func (r *routeRepository) GetNearby(ctx context.Context, lat, lng, distance float64) (*[]Route, error) {
	query := `
        SELECT id, name, description, ST_AsText(path), created_at 
        FROM routes 
        WHERE ST_DWithin(path, ST_SetSRID(ST_MakePoint($1, $2), 4326), $3)
        ORDER BY ST_Distance(path, ST_SetSRID(ST_MakePoint($1, $2), 4326)) ASC
    `

	rows, err := r.db.QueryContext(ctx, query, lng, lat, distance)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes = []Route{}
	for rows.Next() {
		var route Route
		if err := rows.Scan(&route.ID, &route.Name, &route.Description, &route.Path, &route.CreatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	return &routes, nil
}

// Update modifies an existing route, ensuring it belongs to the user
func (r *routeRepository) Update(ctx context.Context, route *Route) error {
	query := `UPDATE routes SET name = $2, description = $3, path = ST_GeomFromText($4, 4326)
              WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, route.ID, route.Name, route.Description, route.Path)
	return err
}

// Delete removes a route, ensuring it belongs to the user
func (r *routeRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM routes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
