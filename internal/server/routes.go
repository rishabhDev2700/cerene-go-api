package server

import (
	"cerene-api/internal/models"
	"context"
	"os"
	"strconv"
	"time"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func (s *FiberServer) RegisterFiberRoutes() {
	s.App.Use(logger.New())
	s.App.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:8081", // Allows requests from any domain
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	// Health check
	s.App.Get("/health", s.healthHandler)

	s.App.Post("/register", s.registerHandler)
	s.App.Post("/login", s.loginHandler)
	secret := os.Getenv("SECRET_KEY")
	s.App.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(secret)},
	}))
	// Route endpoints
	s.App.Post("/routes", s.createRouteHandler)
	s.App.Get("/routes/nearby", s.getNearbyRoutesHandler)
	s.App.Get("/routes/:id", s.getRouteByIDHandler)
	s.App.Put("/routes/:id", s.updateRouteHandler)
	s.App.Delete("/routes/:id", s.deleteRouteHandler)

	// Recommended Stops endpoints
	s.App.Post("/stops", s.createRecommendedStopHandler)
	s.App.Get("/stops/:id", s.getRecommendedStopByIDHandler)
	s.App.Get("/routes/:route_id/stops", s.getStopsByRouteHandler)
	s.App.Put("/stops/:id", s.updateRecommendedStopHandler)
	s.App.Delete("/stops/:id", s.deleteRecommendedStopHandler)
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}

func (s *FiberServer) registerHandler(c *fiber.Ctx) error {
	type RegisterRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	userRepo := models.NewUserRepository(s.db.DB())

	// Check if user already exists
	existingUser, _ := userRepo.GetByEmail(context.Background(), req.Email)
	if existingUser != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email already registered"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Create user
	user := &models.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = userRepo.Create(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered successfully"})
}

// Login Handler
func (s *FiberServer) loginHandler(c *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	userRepo := models.NewUserRepository(s.db.DB())
	// Fetch user by email
	user, err := userRepo.GetByEmail(context.Background(), req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}
	// Create the Claims
	claims := jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

func (s *FiberServer) createRouteHandler(c *fiber.Ctx) error {
	var route models.Route
	if err := c.BodyParser(&route); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := models.NewRouteRepository(s.db.DB()).Create(c.Context(), &route)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create route"})
	}

	return c.Status(fiber.StatusCreated).JSON(route)
}

func (s *FiberServer) getRouteByIDHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid route ID"})
	}

	route, err := models.NewRouteRepository(s.db.DB()).GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch route"})
	}
	if route == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Route not found"})
	}

	return c.JSON(route)
}

func (s *FiberServer) getNearbyRoutesHandler(c *fiber.Ctx) error {
	// Extract latitude and longitude from query parameters
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid latitude"})
	}

	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid longitude"})
	}

	// Extract optional distance parameter (default: 10,000 meters)
	distance, err := strconv.ParseFloat(c.Query("distance", "10000"), 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid distance"})
	}

	// Fetch nearby routes using the repository
	routes, err := models.NewRouteRepository(s.db.DB()).GetNearby(c.Context(), lat, lng, distance)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch nearby routes"})
	}

	return c.JSON(routes)
}

func (s *FiberServer) updateRouteHandler(c *fiber.Ctx) error {
	var route models.Route
	if err := c.BodyParser(&route); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := models.NewRouteRepository(s.db.DB()).Update(c.Context(), &route)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update route"})
	}

	return c.JSON(fiber.Map{"message": "Route updated successfully"})
}

func (s *FiberServer) deleteRouteHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid route ID"})
	}

	err = models.NewRouteRepository(s.db.DB()).Delete(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete route"})
	}

	return c.JSON(fiber.Map{"message": "Route deleted successfully"})
}

func (s *FiberServer) createRecommendedStopHandler(c *fiber.Ctx) error {
	var stop models.RecommendedStop
	if err := c.BodyParser(&stop); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := models.NewRecommendedStopRepository(s.db.DB()).Create(c.Context(), &stop)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(stop)
}

func (s *FiberServer) getRecommendedStopByIDHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid stop ID"})
	}

	stop, err := models.NewRecommendedStopRepository(s.db.DB()).GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch stop"})
	}
	if stop == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Stop not found"})
	}

	return c.JSON(stop)
}

func (s *FiberServer) getStopsByRouteHandler(c *fiber.Ctx) error {
	routeID, err := strconv.Atoi(c.Params("route_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid route ID"})
	}

	stops, err := models.NewRecommendedStopRepository(s.db.DB()).GetByRouteID(c.Context(), routeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch stops"})
	}

	return c.JSON(stops)
}

func (s *FiberServer) updateRecommendedStopHandler(c *fiber.Ctx) error {
	var stop models.RecommendedStop
	if err := c.BodyParser(&stop); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := models.NewRecommendedStopRepository(s.db.DB()).Update(c.Context(), &stop)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update stop"})
	}

	return c.JSON(fiber.Map{"message": "Stop updated successfully"})
}

func (s *FiberServer) deleteRecommendedStopHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid stop ID"})
	}

	err = models.NewRecommendedStopRepository(s.db.DB()).Delete(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete stop"})
	}

	return c.JSON(fiber.Map{"message": "Stop deleted successfully"})
}
