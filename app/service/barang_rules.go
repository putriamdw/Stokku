package service

import (
	"strings"

	"stokku/app/model"
)

// ValidateBarang memeriksa body POST/PUT barang
func ValidateBarang(req model.BarangRequest) map[string]string {
	errs := map[string]string{}
	kode := strings.TrimSpace(req.Kode)
	nama := strings.TrimSpace(req.Nama)
	if kode == "" {
		errs["kode"] = "wajib diisi"
	} else if len(kode) > 30 {
		errs["kode"] = "maksimal 30 karakter"
	}
	if nama == "" {
		errs["nama"] = "wajib diisi"
	} else if len(nama) > 100 {
		errs["nama"] = "maksimal 100 karakter"
	}
	if req.SatuanID < 1 {
		errs["satuan_id"] = "wajib diisi"
	}
	if req.Harga < 0 {
		errs["harga"] = "tidak boleh negatif"
	}
	return errs
}
