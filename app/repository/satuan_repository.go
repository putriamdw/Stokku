package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"stokku/app/model"
)

type SatuanRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Satuan, int, error)
	FindByID(ctx context.Context, id int) (model.Satuan, error)
	Create(ctx context.Context, s model.Satuan) (model.Satuan, error)
	Update(ctx context.Context, s model.Satuan) (model.Satuan, error)
	Deactivate(ctx context.Context, id int) error
}

type satuanRepository struct{ pool *pgxpool.Pool }

func NewSatuanRepository(pool *pgxpool.Pool) SatuanRepository {
	return &satuanRepository{pool: pool}
}

func (r *satuanRepository) FindByID(ctx context.Context, id int) (model.Satuan, error) {
	var s model.Satuan
	err := r.pool.QueryRow(ctx,
		`SELECT id, nama_satuan, is_active FROM satuan WHERE id = $1`, id,
	).Scan(&s.ID, &s.NamaSatuan, &s.IsActive)
	return s, mapError(err)
}

var satuanSortColumns = map[string]string{"id": "id", "nama_satuan": "nama_satuan"}

func (r *satuanRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Satuan, int, error) {
	where, args := "WHERE 1=1", []any{}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND nama_satuan ILIKE $%d", len(args))
	}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND is_active = $%d", len(args))
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM satuan "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	sql := fmt.Sprintf(
		"SELECT id, nama_satuan, is_active FROM satuan %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, sortColumn(satuanSortColumns, q.Sort), orderDir(q.Order), len(args)-1, len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []model.Satuan{}
	for rows.Next() {
		var s model.Satuan
		if err := rows.Scan(&s.ID, &s.NamaSatuan, &s.IsActive); err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, rows.Err()
}

func (r *satuanRepository) Create(ctx context.Context, s model.Satuan) (model.Satuan, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO satuan (nama_satuan) VALUES ($1) RETURNING id, nama_satuan, is_active`,
		s.NamaSatuan,
	).Scan(&s.ID, &s.NamaSatuan, &s.IsActive)
	return s, mapError(err)
}

func (r *satuanRepository) Update(ctx context.Context, s model.Satuan) (model.Satuan, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE satuan SET nama_satuan = $1 WHERE id = $2 RETURNING id, nama_satuan, is_active`,
		s.NamaSatuan, s.ID,
	).Scan(&s.ID, &s.NamaSatuan, &s.IsActive)
	return s, mapError(err)
}

// Deactivate = soft delete. Ditolak (ErrInUse) bila masih dipakai barang aktif
func (r *satuanRepository) Deactivate(ctx context.Context, id int) error {
	if _, err := r.FindByID(ctx, id); err != nil {
		return err
	}
	var used bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM barang WHERE satuan_id = $1 AND is_active)`, id,
	).Scan(&used); err != nil {
		return err
	}
	if used {
		return ErrInUse
	}
	_, err := r.pool.Exec(ctx, `UPDATE satuan SET is_active = FALSE WHERE id = $1`, id)
	return err
}