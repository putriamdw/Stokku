package model

// Satuan adalah satuan barang (pcs, box, kg)
type Satuan struct {
	ID         int    `json:"id"`
	NamaSatuan string `json:"nama_satuan"`
	IsActive   bool   `json:"is_active"`
}

// SatuanRequest adalah body untuk POST dan PUT /satuan
type SatuanRequest struct {
	NamaSatuan string `json:"nama_satuan"`
}