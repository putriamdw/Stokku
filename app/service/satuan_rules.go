package service

import (
	"strings"

	"stokku/app/model"
)

// ValidateSatuan memeriksa body POST/PUT satuan
// Mengembalikan peta field bermasalah; kosong berarti lolos
func ValidateSatuan(req model.SatuanRequest) map[string]string {
	errs := map[string]string{}
	nama := strings.TrimSpace(req.NamaSatuan)
	if nama == "" {
		errs["nama_satuan"] = "wajib diisi"
	} else if len(nama) > 30 {
		errs["nama_satuan"] = "maksimal 30 karakter"
	}
	return errs
}
