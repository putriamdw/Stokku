package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"stokku/app/model"
)

type MarginRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Margin, int, error)
	FindByID(ctx context.Context, id int) (model.Margin, error)
	FindActive(ctx context.Context) (model.Margin, error)
	Create(ctx context.Context, m model.Margin) (model.Margin, error)
	Activate(ctx context.Context, id int) (model.Margin, error)
}

type marginRepository struct{ pool *pgxpool.Pool }

func NewMarginRepository(pool *pgxpool.Pool) MarginRepository {
	return &marginRepository{pool: pool}
}

const marginColumns = `id, persen, is_active, created_by, created_at`

func scanMargin(row scanner) (model.Margin, error) {
	var m model.Margin
	err := row.Scan(&m.ID, &m.Persen, &m.IsActive, &m.CreatedBy, &m.CreatedAt)
	return m, err
}

func (r *marginRepository) FindByID(ctx context.Context, id int) (model.Margin, error) {
	m, err := scanMargin(r.pool.QueryRow(ctx,
		`SELECT `+marginColumns+` FROM margin_penjualan WHERE id = $1`, id))
	return m, mapError(err)
}

// FindActive mengembalikan ErrNotFound bila belum ada margin yang aktif
func (r *marginRepository) FindActive(ctx context.Context) (model.Margin, error) {
	m, err := scanMargin(r.pool.QueryRow(ctx,
		`SELECT `+marginColumns+` FROM margin_penjualan WHERE is_active`))
	return m, mapError(err)
}

var marginSortColumns = map[string]string{"id": "id", "persen": "persen", "created_at": "created_at"}

func (r *marginRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Margin, int, error) {
	where, args := "WHERE 1=1", []any{}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND is_active = $%d", len(args))
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM margin_penjualan "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	sql := fmt.Sprintf("SELECT %s FROM margin_penjualan %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		marginColumns, where, sortColumn(marginSortColumns, q.Sort), orderDir(q.Order),
		len(args)-1, len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []model.Margin{}
	for rows.Next() {
		m, err := scanMargin(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, rows.Err()
}

// Create membuat margin baru dalam keadaan tidak aktif
func (r *marginRepository) Create(ctx context.Context, m model.Margin) (model.Margin, error) {
	out, err := scanMargin(r.pool.QueryRow(ctx,
		`INSERT INTO margin_penjualan (persen, created_by) VALUES ($1, $2) RETURNING `+marginColumns,
		m.Persen, m.CreatedBy))
	return out, mapError(err)
}

// Activate mengaktifkan satu margin dan menonaktifkan yang lain dalam satu transaksi
func (r *marginRepository) Activate(ctx context.Context, id int) (model.Margin, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Margin{}, err
	}
	defer tx.Rollback(ctx) // tidak berpengaruh setelah Commit

	// Kunci agar dua aktivasi bersamaan tidak saling bentrok di unique index
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(7001)`); err != nil {
		return model.Margin{}, err
	}
	if _, err := tx.Exec(ctx,
		`UPDATE margin_penjualan SET is_active = FALSE WHERE is_active AND id <> $1`, id); err != nil {
		return model.Margin{}, err
	}
	m, err := scanMargin(tx.QueryRow(ctx,
		`UPDATE margin_penjualan SET is_active = TRUE WHERE id = $1 RETURNING `+marginColumns, id))
	if err != nil {
		// id tidak ada → ErrNotFound, lalu transaksi di-rollback (margin lama tetap aktif)
		return model.Margin{}, mapError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Margin{}, err
	}
	return m, nil
}