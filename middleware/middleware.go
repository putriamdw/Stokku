package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"stokku/helper"
)

// Register memasang middleware global
func Register(app *fiber.App, logger *slog.Logger) {
	app.Use(requestid.New())       // beri setiap request satu ID unik
	app.Use(recover.New())         // tangkap panic agar server tidak mati
	app.Use(helmet.New())          // header keamanan dasar
	app.Use(cors.New())            // atur CORS
	app.Use(RequestLogger(logger)) // catat setiap request
}

// RequestLogger mencatat setiap request sebagai satu baris JSON
func RequestLogger(logger *slog.Logger) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		requestID, _ := c.Locals("requestid").(string)
		logger.Info("http_request",
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		)
		return err
	}
}

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// RequireJSON menolak request ber-body yang Content-Type-nya bukan JSON
func RequireJSON(c *fiber.Ctx) error {
	if methodsWithBody[c.Method()] {
		if !strings.HasPrefix(c.Get("Content-Type"), fiber.MIMEApplicationJSON) {
			return helper.Fail(c, fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json")
		}
	}
	return c.Next()
}
