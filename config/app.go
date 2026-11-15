package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"stokku/app/service"
	"stokku/helper"
	"stokku/middleware"
	"stokku/route"
)

// NewApp merakit Fiber, middleware, dan route
func NewApp(logger *slog.Logger, pool *pgxpool.Pool, svc *service.Services) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Stokku"),
		ErrorHandler: newErrorResponder(logger),
	})

	middleware.Register(app, logger)
	route.Register(app, pool, svc)

	// Penampung terakhir untuk URL yang tidak dikenal
	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})
	return app
}

// newErrorResponder: jaring pengaman terakhir agar format error tetap konsisten
func newErrorResponder(logger *slog.Logger) func(*fiber.Ctx, error) error {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		message := "terjadi error pada server"
		var fe *fiber.Error
		if errors.As(err, &fe) {
			status = fe.Code
			message = fe.Message
		}
		logger.Error("unhandled_error",
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.String("error", err.Error()),
		)
		return helper.Fail(c, status, message)
	}
}
