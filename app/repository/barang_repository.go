package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"stokku/app/model"
)

type BarangRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Barang, int, error)
	FindByID(ctx context.Context, id int) (model.Barang, error)
	Create(ctx context.Context, b model.Barang) (model.Barang, error)
	Update(ctx context.Context, b model.Barang) (model.Barang, error)
	Deactivate(ctx context.Context, id int) error
}

type barangRepository struct{ pool *pgxpool.Pool }

func NewBarangRepository(pool *pgxpool.Pool) BarangRepository {
	return &barangRepository{pool: pool}
}

const barangSelect = `SELECT b.id, b.kode, b.nama, b.satuan_id, s.nama_satuan,
	b.harga, b.is_active, b.created_at
	FROM barang b JOIN satuan s ON s.id = b.satuan_id`

// scanner dipakai bersama oleh pgx.Row dan pgx.Rows
type scanner interface{ Scan(dest ...any) error }

func scanBarang(row scanner) (model.Barang, error) {
	var b model.Barang
	err := row.Scan(&b.ID, &b.Kode, &b.Nama, &b.SatuanID, &b.NamaSatuan,
		&b.Harga, &b.IsActive, &b.CreatedAt)
	return b, err
}

func (r *barangRepository) FindByID(ctx context.Context, id int) (model.Barang, error) {
	b, err := scanBarang(r.pool.QueryRow(ctx, barangSelect+" WHERE b.id = $1", id))
	return b, mapError(err)
}

var barangSortColumns = map[string]string{
	"id": "b.id", "kode": "b.kode", "nama": "b.nama",
	"harga": "b.harga", "created_at": "b.created_at",
}

func (r *barangRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Barang, int, error) {
	where, args := "WHERE 1=1", []any{}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (b.kode ILIKE $%d OR b.nama ILIKE $%d)", len(args), len(args))
	}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND b.is_active = $%d", len(args))
	}

	var total int
	countSQL := "SELECT COUNT(*) FROM barang b JOIN satuan s ON s.id = b.satuan_id " + where
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	sql := fmt.Sprintf("%s %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		barangSelect, where, sortColumn(barangSortColumns, q.Sort), orderDir(q.Order),
		len(args)-1, len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []model.Barang{}
	for rows.Next() {
		b, err := scanBarang(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, b)
	}
	return items, total, rows.Err()
}

// Create menolak satuan yang tidak ada atau non-aktif (ErrInvalidRef)
func (r *barangRepository) Create(ctx context.Context, b model.Barang) (model.Barang, error) {
	var id int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO barang (kode, nama, satuan_id, harga)
		 SELECT $1::varchar, $2::varchar, s.id, $4::bigint
		 FROM satuan s WHERE s.id = $3 AND s.is_active
		 RETURNING id`,
		b.Kode, b.Nama, b.SatuanID, b.Harga,
	).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Barang{}, ErrInvalidRef
	}
	if err != nil {
		return model.Barang{}, mapError(err)
	}
	return r.FindByID(ctx, id)
}

func (r *barangRepository) Update(ctx context.Context, b model.Barang) (model.Barang, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE barang SET kode = $1, nama = $2, satuan_id = $3, harga = $4
		 WHERE id = $5 AND EXISTS (SELECT 1 FROM satuan WHERE id = $3 AND is_active)`,
		b.Kode, b.Nama, b.SatuanID, b.Harga, b.ID)
	if err != nil {
		return model.Barang{}, mapError(err)
	}
	if tag.RowsAffected() == 0 {
		// Barang tidak ada → ErrNotFound; kalau ada, berarti satuannya yang tidak valid
		if _, err := r.FindByID(ctx, b.ID); err != nil {
			return model.Barang{}, err
		}
		return model.Barang{}, ErrInvalidRef
	}
	return r.FindByID(ctx, b.ID)
}

// Deactivate = soft delete
func (r *barangRepository) Deactivate(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `UPDATE barang SET is_active = FALSE WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}