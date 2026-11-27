package service

import (
	"testing"

	"stokku/app/model"
)

func TestHitungTotal(t *testing.T) {
	cases := []struct {
		name                        string
		items                       []model.DetailPengadaan
		wantSub, wantPpn, wantTotal int64
	}{
		{"kosong", nil, 0, 0, 0},
		{"satu item", []model.DetailPengadaan{{Jumlah: 50, Harga: 2500}}, 125000, 12500, 137500},
		{"dua item", []model.DetailPengadaan{{Jumlah: 10, Harga: 1000}, {Jumlah: 2, Harga: 5000}}, 20000, 2000, 22000},
		{"ppn dibulatkan ke bawah", []model.DetailPengadaan{{Jumlah: 1, Harga: 105}}, 105, 10, 115},
	}
	for _, tc := range cases {
		sub, ppn, total := HitungTotal(tc.items)
		if sub != tc.wantSub || ppn != tc.wantPpn || total != tc.wantTotal {
			t.Errorf("%s: harap (%d,%d,%d), dapat (%d,%d,%d)",
				tc.name, tc.wantSub, tc.wantPpn, tc.wantTotal, sub, ppn, total)
		}
	}
}

func TestValidatePengadaan(t *testing.T) {
	if errs := ValidatePengadaan(model.PengadaanRequest{VendorID: 1, UserID: 1}); len(errs) != 0 {
		t.Errorf("seharusnya lolos, dapat %v", errs)
	}
	errs := ValidatePengadaan(model.PengadaanRequest{})
	for _, f := range []string{"vendor_id", "user_id"} {
		if _, ok := errs[f]; !ok {
			t.Errorf("field %q seharusnya ditolak", f)
		}
	}
}

func TestValidateDetailPengadaan(t *testing.T) {
	cases := []struct {
		name    string
		req     model.DetailPengadaanRequest
		wantErr bool
	}{
		{"valid", model.DetailPengadaanRequest{BarangID: 1, Jumlah: 5}, false},
		{"barang kosong", model.DetailPengadaanRequest{BarangID: 0, Jumlah: 5}, true},
		{"jumlah nol", model.DetailPengadaanRequest{BarangID: 1, Jumlah: 0}, true},
		{"jumlah negatif", model.DetailPengadaanRequest{BarangID: 1, Jumlah: -3}, true},
		{"jumlah terlalu besar", model.DetailPengadaanRequest{BarangID: 1, Jumlah: 1_000_001}, true},
	}
	for _, tc := range cases {
		if got := len(ValidateDetailPengadaan(tc.req)) > 0; got != tc.wantErr {
			t.Errorf("%s: wantErr=%v", tc.name, tc.wantErr)
		}
	}
}