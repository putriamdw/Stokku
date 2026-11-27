package model

import "time"

// margin = persentase keuntungan penjualan, hanya 1 yang boleh aktif
type Margin struct {
	ID        int       `json:"id"`
	Persen    int       `json:"persen"`
	IsActive  bool      `json:"is_active"`
	CreatedBy *int      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// body untuk POST /margin
type MarginRequest struct {
	Persen *int	`json:"persen"`
}