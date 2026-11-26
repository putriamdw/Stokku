package service

import (
	"testing"

	"stokku/app/model"
)

func TestValidateBarang(t *testing.T) {
	valid := model.BarangRequest{Kode: "BRG-001", Nama: "Pulpen", SatuanID: 1, Harga: 2500}
	if errs := ValidateBarang(valid); len(errs) != 0 {
		t.Fatalf("seharusnya lolos, dapat %v", errs)
	}

	bad := model.BarangRequest{Kode: " ", Nama: "", SatuanID: 0, Harga: -1}
	errs := ValidateBarang(bad)
	for _, f := range []string{"kode", "nama", "satuan_id", "harga"} {
		if _, ok := errs[f]; !ok {
			t.Errorf("field %q seharusnya ditolak", f)
		}
	}
}

func TestValidateBarangHargaNol(t *testing.T) {
	req := model.BarangRequest{Kode: "X", Nama: "Gratisan", SatuanID: 1, Harga: 0}
	if errs := ValidateBarang(req); len(errs) != 0 {
		t.Errorf("harga 0 seharusnya diizinkan, dapat %v", errs)
	}
}