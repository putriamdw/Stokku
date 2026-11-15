package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound   = errors.New("data tidak ditemukan")
	ErrDuplicate  = errors.New("data duplikat")
	ErrInUse      = errors.New("data masih dipakai")
	ErrInvalidRef = errors.New("referensi tidak valid atau non-aktif")
)

// mapError menerjemahkan error pgx menjadi sentinel error milik repository
func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicate
	}
	return err
}

// sortColumn memilih kolom ORDER BY dari daftar putih
// (nama kolom tidak bisa di-parameterkan, jadi harus divalidasi)
func sortColumn(allowed map[string]string, key string) string {
	if col, ok := allowed[key]; ok {
		return col
	}
	return allowed["id"]
}

func orderDir(order string) string {
	if order == "desc" {
		return "DESC"
	}
	return "ASC"
}