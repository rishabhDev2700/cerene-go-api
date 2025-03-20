package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// RecommendedStopRepository defines methods for interacting with the recommended_stops table
type RecommendedStopRepository interface {
	Create(ctx context.Context, stop *RecommendedStop) error
	GetByID(ctx context.Context, id int) (*RecommendedStop, error)
	GetByRouteID(ctx context.Context, routeID int) ([]RecommendedStop, error)
	Update(ctx context.Context, stop *RecommendedStop) error
	Delete(ctx context.Context, id int) error
}

// recommendedStopRepository implements RecommendedStopRepository
type recommendedStopRepository struct {
	db *sql.DB
}

// NewRecommendedStopRepository initializes a new RecommendedStopRepository
func NewRecommendedStopRepository(db *sql.DB) RecommendedStopRepository {
	return &recommendedStopRepository{db: db}
}

// Create inserts a new recommended stop into the database
func (r *recommendedStopRepository) Create(ctx context.Context, stop *RecommendedStop) error {
	query := `INSERT INTO recommended_stops (route_id, name, type, description, location, created_at) 
              VALUES ($1, $2, $3, $4, ST_GeomFromText($5, 4326), NOW())`
	_, err := r.db.ExecContext(ctx, query, stop.RouteID, stop.Name, stop.Type, stop.Description, stop.Location)

	if err != nil {
		return fmt.Errorf("failed to insert recommended stop: %w", err)
	}

	return nil
}

// GetByID fetches a recommended stop by its ID
func (r *recommendedStopRepository) GetByID(ctx context.Context, id int) (*RecommendedStop, error) {
	query := `SELECT id, route_id, name, type, description, ST_AsText(location), created_at 
              FROM recommended_stops WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var stop RecommendedStop
	err := row.Scan(&stop.ID, &stop.RouteID, &stop.Name, &stop.Type, &stop.Description, &stop.Location, &stop.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &stop, nil
}

// GetByRouteID fetches all recommended stops for a given route
func (r *recommendedStopRepository) GetByRouteID(ctx context.Context, routeID int) ([]RecommendedStop, error) {
	query := `SELECT id, route_id, name, type, description, ST_AsText(location), created_at 
              FROM recommended_stops WHERE route_id = $1`
	rows, err := r.db.QueryContext(ctx, query, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stops []RecommendedStop
	for rows.Next() {
		var stop RecommendedStop
		if err := rows.Scan(&stop.ID, &stop.RouteID, &stop.Name, &stop.Type, &stop.Description, &stop.Location, &stop.CreatedAt); err != nil {
			return nil, err
		}
		stops = append(stops, stop)
	}

	return stops, nil
}

// Update modifies an existing recommended stop
func (r *recommendedStopRepository) Update(ctx context.Context, stop *RecommendedStop) error {
	query := `UPDATE recommended_stops 
              SET name = $2, type = $3, description = $4, location = ST_GeomFromText($5, 4326) 
              WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, stop.ID, stop.Name, stop.Type, stop.Description, stop.Location)
	return err
}

// Delete removes a recommended stop by its ID
func (r *recommendedStopRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM recommended_stops WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
