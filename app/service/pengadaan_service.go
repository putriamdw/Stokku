package service

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"stokku/app/model"
	"stokku/app/repository"
	"stokku/helper"
)

type PengadaanService struct {
	repo repository.PengadaanRepository
}

func NewPengadaanService(repo repository.PengadaanRepository) *PengadaanService {
	return &PengadaanService{repo: repo}
}

var pengadaanSortable = map[string]bool{"id": true, "tanggal": true, "total": true}

func (s *PengadaanService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	q := helper.ParseListQuery(c, pengadaanSortable)
	status := c.Query("status")
	if status != "" && !IsValidStatusPengadaan(status) {
		return helper.Fail(c, fiber.StatusBadRequest, "status harus diproses atau selesai")
	}
	items, total, err := s.repo.FindAll(ctx, q, status)
	if err != nil {
		return translateError(c, err, "pengadaan", "gagal mengambil data pengadaan")
	}
	return helper.SuccessList(c, "daftar pengadaan berhasil diambil", items, &model.Meta{
		Page: q.Page, Limit: q.Limit, Total: total,
		TotalPages: CountTotalPages(total, q.Limit),
	})
}

// Unfinished: pengadaan yang masih punya sisa barang (dipakai form penerimaan).
func (s *PengadaanService) Unfinished(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	items, err := s.repo.FindUnfinished(ctx)
	if err != nil {
		return translateError(c, err, "pengadaan", "gagal mengambil data pengadaan")
	}
	return helper.Success(c, fiber.StatusOK, "pengadaan yang belum selesai berhasil diambil", items)
}

func (s *PengadaanService) Get(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateError(c, err, "pengadaan", "gagal mengambil data pengadaan")
	}
	return helper.Success(c, fiber.StatusOK, "pengadaan ditemukan", item)
}

func (s *PengadaanService) Create(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	var req model.PengadaanRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := ValidatePengadaan(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	item, err := s.repo.Create(ctx, model.Pengadaan{VendorID: req.VendorID, UserID: req.UserID})
	if err != nil {
		return translateError(c, err, "pengadaan", "gagal menyimpan pengadaan")
	}
	return helper.Created(c, "pengadaan berhasil dibuat", item,
		"/api/v1/pengadaan/"+strconv.Itoa(item.ID))
}

func (s *PengadaanService) AddDetail(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	var req model.DetailPengadaanRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := ValidateDetailPengadaan(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	// HitungTotal (aturan bisnis) disuntik ke repository agar dijalankan di dalam transaksi.
	item, err := s.repo.AddDetail(ctx, id, req.BarangID, req.Jumlah, HitungTotal)
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Fail(c, fiber.StatusConflict, "barang sudah ada pada pengadaan ini")
	}
	if err != nil {
		return translateError(c, err, "pengadaan", "gagal menambah barang ke pengadaan")
	}
	return helper.Created(c, "barang berhasil ditambahkan ke pengadaan", item,
		"/api/v1/pengadaan/"+strconv.Itoa(id))
}

func (s *PengadaanService) DeleteDetail(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamInt(c, "id")
	detailID, ok2 := helper.ParamInt(c, "detail_id")
	if !ok || !ok2 {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	item, err := s.repo.DeleteDetail(ctx, id, detailID, HitungTotal)
	if errors.Is(err, repository.ErrNotFound) {
		return helper.Fail(c, fiber.StatusNotFound, "pengadaan atau detail tidak ditemukan")
	}
	if err != nil {
		return translateError(c, err, "pengadaan", "gagal menghapus detail pengadaan")
	}
	return helper.Success(c, fiber.StatusOK, "detail berhasil dihapus", item)
}

func (s *PengadaanService) Delete(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	id, ok := helper.ParamID(c)
	if !ok {
		return helper.Fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return translateError(c, err, "pengadaan", "gagal menghapus pengadaan")
	}
	return helper.NoContent(c)
}