package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"stokku/app/model"
	"stokku/app/repository"
	"stokku/helper"
)

type MarginService struct {
	repo repository.MarginRepository
}

func NewMarginService(repo repository.MarginRepository) *MarginService {
	return &MarginService{repo: repo}
}

var marginSortable = map[string]bool{"id": true, "persen": true, "created_at": true}

func (s *MarginService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	q := helper.ParseListQuery(c, marginSortable)
	items, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return translateError(c, err, "margin", "gagal mengambil data margin")
	}
	return helper.SuccessList(c, "daftar margin berhasil diambil", items, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *MarginService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "margin", "gagal mengambil data margin")
	}
	return helper.Success(c, fiber.StatusOK, "margin ditemukan", item)
}

// Active mengambil margin yang sedang aktif (dipakai form penjualan)
func (s *MarginService) Active(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	item, err := s.repo.FindActive(ctx)
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "belum ada margin aktif")
	}
	if err != nil {
		return translateError(c, err, "margin", "gagal mengambil margin aktif")
	}
	return helper.Success(c, fiber.StatusOK, "margin aktif ditemukan", item)
}

func (s *MarginService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	var req model.MarginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := ValidateMargin(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	
	item, err := s.repo.Create(ctx, model.Margin{Persen: *req.Persen})
	if err != nil {
		return translateError(c, err, "margin", "gagal menyimpan margin")
	}
	return helper.Created(c, "margin berhasil dibuat", item,
		"/api/v1/margin/"+strconv.Itoa(item.ID))
}

// Activate menjadikan satu margin aktif dan menonaktifkan margin lainnya
func (s *MarginService) Activate(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	item, err := s.repo.Activate(ctx, id)
	if err != nil {
		return translateError(c, err, "margin", "gagal mengaktifkan margin")
	}
	return helper.Success(c, fiber.StatusOK, "margin berhasil diaktifkan", item)
}