package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"stokku/app/service"
	"stokku/helper"
	"stokku/middleware"
)

// Register memetakan URL ke method service
func Register(app *fiber.App, pool *pgxpool.Pool, svc *service.Services) {
	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(pool))

	satuan := api.Group("/satuan", middleware.RequireJSON)
	satuan.Get("/", svc.Satuan.List)
	satuan.Get("/:id", svc.Satuan.Get)
	satuan.Post("/", svc.Satuan.Create)
	satuan.Put("/:id", svc.Satuan.Replace)
	satuan.Delete("/:id", svc.Satuan.Delete)

	barang := api.Group("/barang", middleware.RequireJSON)
	barang.Get("/", svc.Barang.List)
	barang.Get("/:id", svc.Barang.Get)
	barang.Post("/", svc.Barang.Create)
	barang.Put("/:id", svc.Barang.Replace)
	barang.Delete("/:id", svc.Barang.Delete)
}

func healthCheck(pool *pgxpool.Pool) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
