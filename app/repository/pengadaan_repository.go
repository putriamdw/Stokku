package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"stokku/app/model"
)

// TotalFunc menghitung subtotal, ppn, total dari daftar detail
// Aturan hitungnya milik service, repository hanya memanggilnya di dalam transaksi
type TotalFunc func(items []model.DetailPengadaan) (subtotal, ppn, total int64)

type PengadaanRepository interface {
	FindAll(ctx context.Context, q model.ListQuery, status string) ([]model.Pengadaan, int, error)
	FindByID(ctx context.Context, id int) (model.PengadaanLengkap, error)
	FindUnfinished(ctx context.Context) ([]model.Pengadaan, error)
	Create(ctx context.Context, p model.Pengadaan) (model.Pengadaan, error)
	AddDetail(ctx context.Context, pengadaanID, barangID, jumlah int, hitung TotalFunc) (model.PengadaanLengkap, error)
	DeleteDetail(ctx context.Context, pengadaanID, detailID int, hitung TotalFunc) (model.PengadaanLengkap, error)
	Delete(ctx context.Context, id int) error
}

type pengadaanRepository struct{ pool *pgxpool.Pool }

func NewPengadaanRepository(pool *pgxpool.Pool) PengadaanRepository {
	return &pengadaanRepository{pool: pool}
}

// queryer dipenuhi oleh pgxpool.Pool maupun pgx.Tx,
// sehingga helper baca bisa dipakai di dalam dan di luar transaksi
type queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

const pengadaanSelect = `SELECT p.id, p.vendor_id, v.nama, p.user_id, p.tanggal,
	p.status, p.subtotal, p.ppn, p.total
	FROM pengadaan p JOIN vendor v ON v.id = p.vendor_id`

func scanPengadaan(row scanner) (model.Pengadaan, error) {
	var p model.Pengadaan
	err := row.Scan(&p.ID, &p.VendorID, &p.NamaVendor, &p.UserID, &p.Tanggal,
		&p.Status, &p.Subtotal, &p.Ppn, &p.Total)
	return p, err
}

func findHeader(ctx context.Context, db queryer, id int) (model.Pengadaan, error) {
	p, err := scanPengadaan(db.QueryRow(ctx, pengadaanSelect+" WHERE p.id = $1", id))
	return p, mapError(err)
}

// Create menolak vendor non-aktif/tidak ada atau user tidak ada (ErrInvalidRef)
func (r *pengadaanRepository) Create(ctx context.Context, p model.Pengadaan) (model.Pengadaan, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO pengadaan (vendor_id, user_id)
		 SELECT v.id, u.id FROM vendor v, users u
		 WHERE v.id = $1 AND v.is_active AND u.id = $2
		 RETURNING id`,
		p.VendorID, p.UserID,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Pengadaan{}, ErrInvalidRef
	}
	if err != nil {
		return model.Pengadaan{}, mapError(err)
	}
	return findHeader(ctx, r.pool, id)
}

// findItems mengambil detail beserta jumlah yang sudah diterima dan sisanya.
func findItems(ctx context.Context, db queryer, pengadaanID int) ([]model.DetailPengadaan, error) {
	rows, err := db.Query(ctx,
		`SELECT d.id, d.pengadaan_id, d.barang_id, b.kode, b.nama, d.jumlah, d.harga, d.subtotal,
		        COALESCE((SELECT SUM(dp.jumlah_terima)
		                  FROM detail_penerimaan dp
		                  JOIN penerimaan pn ON pn.id = dp.penerimaan_id
		                  WHERE pn.pengadaan_id = d.pengadaan_id AND dp.barang_id = d.barang_id), 0) AS diterima
		 FROM detail_pengadaan d JOIN barang b ON b.id = d.barang_id
		 WHERE d.pengadaan_id = $1 ORDER BY d.id`, pengadaanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.DetailPengadaan{}
	for rows.Next() {
		var it model.DetailPengadaan
		if err := rows.Scan(&it.ID, &it.PengadaanID, &it.BarangID, &it.KodeBarang, &it.NamaBarang,
			&it.Jumlah, &it.Harga, &it.Subtotal, &it.Diterima); err != nil {
			return nil, err
		}
		it.Sisa = it.Jumlah - it.Diterima
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *pengadaanRepository) FindByID(ctx context.Context, id int) (model.PengadaanLengkap, error) {
	head, err := findHeader(ctx, r.pool, id)
	if err != nil {
		return model.PengadaanLengkap{}, err
	}
	items, err := findItems(ctx, r.pool, id)
	if err != nil {
		return model.PengadaanLengkap{}, err
	}
	return model.PengadaanLengkap{Pengadaan: head, Items: items}, nil
}

var pengadaanSortColumns = map[string]string{
	"id": "p.id", "tanggal": "p.tanggal", "total": "p.total",
}

func (r *pengadaanRepository) FindAll(ctx context.Context, q model.ListQuery, status string) ([]model.Pengadaan, int, error) {
	where, args := "WHERE 1=1", []any{}
	if status != "" {
		args = append(args, status)
		where += fmt.Sprintf(" AND p.status = $%d", len(args))
	}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND v.nama ILIKE $%d", len(args))
	}

	var total int
	countSQL := "SELECT COUNT(*) FROM pengadaan p JOIN vendor v ON v.id = p.vendor_id " + where
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	sql := fmt.Sprintf("%s %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		pengadaanSelect, where, sortColumn(pengadaanSortColumns, q.Sort), orderDir(q.Order),
		len(args)-1, len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []model.Pengadaan{}
	for rows.Next() {
		p, err := scanPengadaan(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, p)
	}
	return items, total, rows.Err()
}

// FindUnfinished: pengadaan diproses yang masih punya sisa barang (untuk form penerimaan)
func (r *pengadaanRepository) FindUnfinished(ctx context.Context) ([]model.Pengadaan, error) {
	rows, err := r.pool.Query(ctx, pengadaanSelect+`
		WHERE p.status = 'diproses' AND EXISTS (
			SELECT 1 FROM detail_pengadaan d
			WHERE d.pengadaan_id = p.id AND d.jumlah > COALESCE((
				SELECT SUM(dp.jumlah_terima)
				FROM detail_penerimaan dp
				JOIN penerimaan pn ON pn.id = dp.penerimaan_id
				WHERE pn.pengadaan_id = d.pengadaan_id AND dp.barang_id = d.barang_id), 0))
		ORDER BY p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.Pengadaan{}
	for rows.Next() {
		p, err := scanPengadaan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

// lockEditable mengunci baris pengadaan lalu memastikan belum ada penerimaan.
// Penerimaan juga akan mengunci baris yang sama, sehingga keduanya tidak bisa bentrok.
func lockEditable(ctx context.Context, tx pgx.Tx, id int) error {
	var locked int
	if err := tx.QueryRow(ctx,
		`SELECT id FROM pengadaan WHERE id = $1 FOR UPDATE`, id).Scan(&locked); err != nil {
		return mapError(err)
	}
	var ada bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM penerimaan WHERE pengadaan_id = $1)`, id).Scan(&ada); err != nil {
		return err
	}
	if ada {
		return ErrLocked
	}
	return nil
}

// recalc menghitung ulang total dengan callback dari service lalu menyimpannya.
func recalc(ctx context.Context, tx pgx.Tx, id int, hitung TotalFunc) error {
	items, err := findItems(ctx, tx, id)
	if err != nil {
		return err
	}
	subtotal, ppn, total := hitung(items)
	_, err = tx.Exec(ctx,
		`UPDATE pengadaan SET subtotal = $1, ppn = $2, total = $3 WHERE id = $4`,
		subtotal, ppn, total, id)
	return err
}

// AddDetail: harga diambil dari barang (snapshot). Duplikat barang → ErrDuplicate,
// barang tidak ada/non-aktif → ErrInvalidRef, sudah ada penerimaan → ErrLocked.
func (r *pengadaanRepository) AddDetail(ctx context.Context, pengadaanID, barangID, jumlah int, hitung TotalFunc) (model.PengadaanLengkap, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.PengadaanLengkap{}, err
	}
	defer tx.Rollback(ctx) // tidak berpengaruh setelah Commit

	if err := lockEditable(ctx, tx, pengadaanID); err != nil {
		return model.PengadaanLengkap{}, err
	}
	tag, err := tx.Exec(ctx,
		`INSERT INTO detail_pengadaan (pengadaan_id, barang_id, jumlah, harga, subtotal)
		 SELECT $1::int, b.id, $3::int, b.harga, $3::int * b.harga
		 FROM barang b WHERE b.id = $2 AND b.is_active`,
		pengadaanID, barangID, jumlah)
	if err != nil {
		return model.PengadaanLengkap{}, mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return model.PengadaanLengkap{}, ErrInvalidRef
	}
	if err := recalc(ctx, tx, pengadaanID, hitung); err != nil {
		return model.PengadaanLengkap{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.PengadaanLengkap{}, err
	}
	return r.FindByID(ctx, pengadaanID)
}

func (r *pengadaanRepository) DeleteDetail(ctx context.Context, pengadaanID, detailID int, hitung TotalFunc) (model.PengadaanLengkap, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.PengadaanLengkap{}, err
	}
	defer tx.Rollback(ctx)

	if err := lockEditable(ctx, tx, pengadaanID); err != nil {
		return model.PengadaanLengkap{}, err
	}
	tag, err := tx.Exec(ctx,
		`DELETE FROM detail_pengadaan WHERE id = $1 AND pengadaan_id = $2`, detailID, pengadaanID)
	if err != nil {
		return model.PengadaanLengkap{}, err
	}
	if tag.RowsAffected() == 0 {
		return model.PengadaanLengkap{}, ErrNotFound
	}
	if err := recalc(ctx, tx, pengadaanID, hitung); err != nil {
		return model.PengadaanLengkap{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.PengadaanLengkap{}, err
	}
	return r.FindByID(ctx, pengadaanID)
}

// Delete menghapus PO beserta detailnya (ON DELETE CASCADE), hanya jika belum ada penerimaan.
func (r *pengadaanRepository) Delete(ctx context.Context, id int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := lockEditable(ctx, tx, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM pengadaan WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
