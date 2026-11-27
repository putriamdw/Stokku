package service

import "stokku/app/model"

// memeriksa body POST margin (persen 0-100)
func ValidateMargin(req model.MarginRequest) map[string]string {
	errs := map[string]string{}
	switch {
	case req.Persen == nil:
		errs["persen"] = "wajib diisi"
	case *req.Persen < 0:
		errs["persen"] = "tidak boleh negatif"
	case *req.Persen > 1000:
		errs["persen"] = "maksimal 1000"
	}
	return errs
}