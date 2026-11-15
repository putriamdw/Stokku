package service

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"stokku/app/model"
	"stokku/app/repository"
	"stokku/helper"
)

type BarangService struct {
	repo repository.BarangRepository
}

func NewBarangService(repo repository.BarangRepository) *BarangService {
	return &BarangService{repo: repo}
}

var barangSortable = map[string]bool{
	"id": true, "kode": true, "nama": true, "harga": true, "created_at": true,
}

func (s *BarangService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	q := helper.ParseListQuery(c, barangSortable)
	items, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return translateError(c, err, "barang", "gagal mengambil data barang")
	}
	return helper.SuccessList(c, "daftar barang berhasil diambil", items, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *BarangService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "barang", "gagal mengambil data barang")
	}
	return helper.Success(c, fiber.StatusOK, "barang ditemukan", item)
}

func (s *BarangService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	var req model.BarangRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Kode, req.Nama = strings.TrimSpace(req.Kode), strings.TrimSpace(req.Nama)
	if errs := ValidateBarang(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	item, err := s.repo.Create(ctx, model.Barang{
		Kode: req.Kode, Nama: req.Nama, SatuanID: req.SatuanID, Harga: req.Harga,
	})
	if err != nil {
		return translateError(c, err, "barang", "gagal menyimpan barang")
	}
	return helper.Created(c, "barang berhasil dibuat", item,
		"/api/v1/barang/"+strconv.Itoa(item.ID))
}

func (s *BarangService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	var req model.BarangRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Kode, req.Nama = strings.TrimSpace(req.Kode), strings.TrimSpace(req.Nama)
	if errs := ValidateBarang(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	item, err := s.repo.Update(ctx, model.Barang{
		ID: id, Kode: req.Kode, Nama: req.Nama, SatuanID: req.SatuanID, Harga: req.Harga,
	})
	if err != nil {
		return translateError(c, err, "barang", "gagal memperbarui barang")
	}
	return helper.Success(c, fiber.StatusOK, "barang berhasil diperbarui", item)
}

func (s *BarangService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	if err := s.repo.Deactivate(ctx, id); err != nil {
		return translateError(c, err, "barang", "gagal menonaktifkan barang")
	}
	return helper.NoContent(c)
}
