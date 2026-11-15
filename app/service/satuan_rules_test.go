package service

import (
	"strings"
	"testing"

	"stokku/app/model"
)

func TestValidateSatuan(t *testing.T) {
	cases := []struct {
		name    string
		nama    string
		wantErr bool
	}{
		{"valid", "pcs", false},
		{"kosong", "", true},
		{"spasi saja", "   ", true},
		{"terlalu panjang", strings.Repeat("a", 31), true},
	}
	for _, tc := range cases {
		errs := ValidateSatuan(model.SatuanRequest{NamaSatuan: tc.nama})
		if got := len(errs) > 0; got != tc.wantErr {
			t.Errorf("%s: wantErr=%v, dapat %v", tc.name, tc.wantErr, errs)
		}
	}
}
