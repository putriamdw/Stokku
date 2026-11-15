package service

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"stokku/app/repository"
	"stokku/helper"
)

// translateError memetakan error repository menjadi status HTTP
func translateError(c *fiber.Ctx, err error, resource, general string) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(c, fiber.StatusNotFound, resource+" tidak ditemukan")
	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(c, fiber.StatusConflict, resource+" sudah ada")
	case errors.Is(err, repository.ErrInUse):
		return helper.Fail(c, fiber.StatusUnprocessableEntity,
			resource+" masih dipakai, tidak bisa dinonaktifkan")
	case errors.Is(err, repository.ErrInvalidRef):
		return helper.Fail(c, fiber.StatusUnprocessableEntity,
			"referensi tidak valid atau non-aktif")
	default:
		slog.Error("service_error",
			slog.String("resource", resource), slog.String("error", err.Error()))
		return helper.Fail(c, fiber.StatusInternalServerError, general)
	}
}
