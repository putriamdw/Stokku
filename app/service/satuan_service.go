package service

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"stokku/app/model"
	"stokku/app/repository"
	"stokku/helper"
)

// SatuanService: menerima fiber.Ctx (peran penghubung HTTP) dan
// memanggil rules (peran use case)
type SatuanService struct {
	repo repository.SatuanRepository
}

// NewSatuanService menerima INTERFACE, bukan struct konkret
func NewSatuanService(repo repository.SatuanRepository) *SatuanService {
	return &SatuanService{repo: repo}
}

var satuanSortable = map[string]bool{"id": true, "nama_satuan": true}

func (s *SatuanService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	q := helper.ParseListQuery(c, satuanSortable)
	items, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return translateError(c, err, "satuan", "gagal mengambil data satuan")
	}
	return helper.SuccessList(c, "daftar satuan berhasil diambil", items, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

func (s *SatuanService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "satuan", "gagal mengambil data satuan")
	}
	return helper.Success(c, fiber.StatusOK, "satuan ditemukan", item)
}

func (s *SatuanService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	var req model.SatuanRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.NamaSatuan = strings.TrimSpace(req.NamaSatuan)
	if errs := ValidateSatuan(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	item, err := s.repo.Create(ctx, model.Satuan{NamaSatuan: req.NamaSatuan})
	if err != nil {
		return translateError(c, err, "satuan", "gagal menyimpan satuan")
	}
	return helper.Created(c, "satuan berhasil dibuat", item,
		"/api/v1/satuan/"+strconv.Itoa(item.ID))
}

func (s *SatuanService) Replace(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	var req model.SatuanRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.NamaSatuan = strings.TrimSpace(req.NamaSatuan)
	if errs := ValidateSatuan(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	item, err := s.repo.Update(ctx, model.Satuan{ID: id, NamaSatuan: req.NamaSatuan})
	if err != nil {
		return translateError(c, err, "satuan", "gagal memperbarui satuan")
	}
	return helper.Success(c, fiber.StatusOK, "satuan berhasil diperbarui", item)
}

func (s *SatuanService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	if err := s.repo.Deactivate(ctx, id); err != nil {
		return translateError(c, err, "satuan", "gagal menonaktifkan satuan")
	}
	return helper.NoContent(c)
}
