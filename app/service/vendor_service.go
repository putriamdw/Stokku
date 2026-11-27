package service

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"stokku/app/model"
	"stokku/app/repository"
	"stokku/helper"
)

type VendorService struct {
	repo repository.VendorRepository
}

func NewVendorService(repo repository.VendorRepository) *VendorService {
	return &VendorService{repo: repo}
}

var vendorSortable = map[string]bool{"id": true, "nama": true, "created_at": true}

// trimVendor merapikan spasi di tepi semua field teks
func trimVendor(req model.VendorRequest) model.VendorRequest {
	return model.VendorRequest{
		Nama:    strings.TrimSpace(req.Nama),
		Alamat:  strings.TrimSpace(req.Alamat),
		Telepon: strings.TrimSpace(req.Telepon),
	}
}

func (s *VendorService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	q := helper.ParseListQuery(c, vendorSortable)
	items, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return translateError(c, err, "vendor", "gagal mengambil data vendor")
	}
	return helper.SuccessList(c, "daftar vendor berhasil diambil", items, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *VendorService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "vendor", "gagal mengambil data vendor")
	}
	return helper.Success(c, fiber.StatusOK, "vendor ditemukan", item)
}

func (s *VendorService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	var req model.VendorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req = trimVendor(req)
	if errs := ValidateVendor(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	item, err := s.repo.Create(ctx, model.Vendor{
		Nama: req.Nama, Alamat: req.Alamat, Telepon: req.Telepon,
	})
	if err != nil {
		return translateError(c, err, "vendor", "gagal menyimpan vendor")
	}
	return helper.Created(c, "vendor berhasil dibuat", item,
		"/api/v1/vendor/"+strconv.Itoa(item.ID))
}

func (s *VendorService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	var req model.VendorRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req = trimVendor(req)
	if errs := ValidateVendor(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	item, err := s.repo.Update(ctx, model.Vendor{
		ID: id, Nama: req.Nama, Alamat: req.Alamat, Telepon: req.Telepon,
	})
	if err != nil {
		return translateError(c, err, "vendor", "gagal memperbarui vendor")
	}
	return helper.Success(c, fiber.StatusOK, "vendor berhasil diperbarui", item)
}

func (s *VendorService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	if err := s.repo.Deactivate(ctx, id); err != nil {
		return translateError(c, err, "vendor", "gagal menonaktifkan vendor")
	}
	return helper.NoContent(c)
}