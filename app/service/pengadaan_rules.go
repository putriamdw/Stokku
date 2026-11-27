package service

import "stokku/app/model"

// ppnPersen adalah tarif PPN pengadaan (10 persen)
const ppnPersen = 10

// maxJumlahItem membatasi jumlah per item agar perkalian tidak meluap
const maxJumlahItem = 1_000_000

// ValidatePengadaan memeriksa body POST /pengadaan
func ValidatePengadaan(req model.PengadaanRequest) map[string]string {
	errs := map[string]string{}
	if req.VendorID < 1 {
		errs["vendor_id"] = "wajib diisi"
	}
	if req.UserID < 1 {
		errs["user_id"] = "wajib diisi"
	}
	return errs
}

// ValidateDetailPengadaan memeriksa body POST /pengadaan/:id/detail
func ValidateDetailPengadaan(req model.DetailPengadaanRequest) map[string]string {
	errs := map[string]string{}
	if req.BarangID < 1 {
		errs["barang_id"] = "wajib diisi"
	}
	if req.Jumlah < 1 {
		errs["jumlah"] = "harus lebih dari 0"
	} else if req.Jumlah > maxJumlahItem {
		errs["jumlah"] = "terlalu besar"
	}
	return errs
}

// HitungTotal menghitung subtotal = jumlah x harga tiap item,
// PPN = subtotal x 10 / 100 (bilangan bulat, dibulatkan ke bawah), total = subtotal + PPN
func HitungTotal(items []model.DetailPengadaan) (subtotal, ppn, total int64) {
	for _, it := range items {
		subtotal += int64(it.Jumlah) * it.Harga
	}
	ppn = subtotal * ppnPersen / 100
	return subtotal, ppn, subtotal + ppn
}

// IsValidStatusPengadaan memeriksa nilai filter ?status= pada list pengadaan
func IsValidStatusPengadaan(status string) bool {
	return status == "diproses" || status == "selesai"
}