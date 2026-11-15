package model

import "time"

// Barang adalah data master barang. Harga = harga beli (rupiah)
type Barang struct {
	ID         int       `json:"id"`
	Kode       string    `json:"kode"`
	Nama       string    `json:"nama"`
	SatuanID   int       `json:"satuan_id"`
	NamaSatuan string    `json:"nama_satuan"`
	Harga      int64     `json:"harga"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

// BarangRequest adalah body untuk POST dan PUT /barang
type BarangRequest struct {
	Kode     string `json:"kode"`
	Nama     string `json:"nama"`
	SatuanID int    `json:"satuan_id"`
	Harga    int64  `json:"harga"`
}