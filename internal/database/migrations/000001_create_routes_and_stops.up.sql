-- Enable PostGIS extension (if not already enabled)
CREATE EXTENSION IF NOT EXISTS postgis;

-- Create routes table
CREATE TABLE routes (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    path GEOMETRY(LINESTRING, 4326) NOT NULL, -- Store path as a LINESTRING
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create recommended_stops table
CREATE TABLE recommended_stops (
    id SERIAL PRIMARY KEY,
    route_id INT REFERENCES routes(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    type TEXT CHECK (type IN ('viewpoint', 'shop', 'landmark', 'other')),
    description TEXT,
    location GEOMETRY(POINT, 4326) NOT NULL, -- Store location as a POINT
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
