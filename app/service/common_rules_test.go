package service

import "testing"

func TestCountTotalPages(t *testing.T) {
	cases := []struct{ total, limit, want int }{
		{0, 10, 0}, {1, 10, 1}, {10, 10, 1}, {11, 10, 2}, {137, 20, 7},
	}
	for _, tc := range cases {
		if got := CountTotalPages(tc.total, tc.limit); got != tc.want {
			t.Errorf("total=%d limit=%d: harap %d, dapat %d", tc.total, tc.limit, tc.want, got)
		}
	}
}
