package service

import (
	"regexp"
	"strings"

	"stokku/app/model"
)

var teleponValid = regexp.MustCompile(`^[0-9+() -]+$`)

// ValidateVendor memeriksa body POST/PUT vendor
// Mengembalikan peta field bermasalah, kosong berarti lolos
func ValidateVendor(req model.VendorRequest) map[string]string {
	errs := map[string]string{}

	nama := strings.TrimSpace(req.Nama)
	if nama == "" {
		errs["nama"] = "wajib diisi"
	} else if len(nama) > 100 {
		errs["nama"] = "maksimal 100 karakter"
	}

	if len(strings.TrimSpace(req.Alamat)) > 255 {
		errs["alamat"] = "maksimal 255 karakter"
	}

	telepon := strings.TrimSpace(req.Telepon)
	if telepon != "" {
		if len(telepon) > 20 {
			errs["telepon"] = "maksimal 20 karakter"
		} else if !teleponValid.MatchString(telepon) {
			errs["telepon"] = "hanya boleh angka, spasi, +, -, dan tanda kurung"
		}
	}
	return errs
}