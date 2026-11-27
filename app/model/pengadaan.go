package model

import "time"

// Pengadaan adalah header PO ke vendor. Subtotal, PPN, dan total dihitung backend
type Pengadaan struct {
	ID         int       `json:"id"`
	VendorID   int       `json:"vendor_id"`
	NamaVendor string    `json:"nama_vendor"`
	UserID     int       `json:"user_id"`
	Tanggal    time.Time `json:"tanggal"`
	Status     string    `json:"status"`
	Subtotal   int64     `json:"subtotal"`
	Ppn        int64     `json:"ppn"`
	Total      int64     `json:"total"`
}

// DetailPengadaan adalah satu baris barang pada PO
// Diterima dan Sisa dihitung dari tabel penerimaan (hanya terisi pada GET detail)
type DetailPengadaan struct {
	ID          int    `json:"id"`
	PengadaanID int    `json:"pengadaan_id"`
	BarangID    int    `json:"barang_id"`
	KodeBarang  string `json:"kode_barang"`
	NamaBarang  string `json:"nama_barang"`
	Jumlah      int    `json:"jumlah"`
	Harga       int64  `json:"harga"`
	Subtotal    int64  `json:"subtotal"`
	Diterima    int    `json:"diterima"`
	Sisa        int    `json:"sisa"`
}

// PengadaanLengkap = header + daftar detail (untuk GET /pengadaan/:id).
type PengadaanLengkap struct {
	Pengadaan
	Items []DetailPengadaan `json:"items"`
}

// PengadaanRequest adalah body POST /pengadaan.
// user_id dikirim manual selama belum ada login; nanti diambil dari token.
type PengadaanRequest struct {
	VendorID int `json:"vendor_id"`
	UserID   int `json:"user_id"`
}

// DetailPengadaanRequest adalah body POST /pengadaan/:id/detail.
type DetailPengadaanRequest struct {
	BarangID int `json:"barang_id"`
	Jumlah   int `json:"jumlah"`
}