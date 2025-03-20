package server

import (
	"github.com/gofiber/fiber/v2"

	"cerene-api/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "cerene-api",
			AppName:      "cerene-api",
		}),

		db: database.New(),
	}
	server.App.Hooks().OnShutdown(func() error {
		return server.db.Close()
	})
	return server
}
