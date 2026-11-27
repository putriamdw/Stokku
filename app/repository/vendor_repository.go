package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"stokku/app/model"
)

type VendorRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Vendor, int, error)
	FindByID(ctx context.Context, id int) (model.Vendor, error)
	Create(ctx context.Context, v model.Vendor) (model.Vendor, error)
	Update(ctx context.Context, v model.Vendor) (model.Vendor, error)
	Deactivate(ctx context.Context, id int) error
}

type vendorRepository struct{ pool *pgxpool.Pool }

func NewVendorRepository(pool *pgxpool.Pool) VendorRepository {
	return &vendorRepository{pool: pool}
}

// alamat dan telepon boleh NULL di database
const vendorColumns = `id, nama, COALESCE(alamat, ''), COALESCE(telepon, ''), is_active, created_at`

func scanVendor(row scanner) (model.Vendor, error) {
	var v model.Vendor
	err := row.Scan(&v.ID, &v.Nama, &v.Alamat, &v.Telepon, &v.IsActive, &v.CreatedAt)
	return v, err
}

func (r *vendorRepository) FindByID(ctx context.Context, id int) (model.Vendor, error) {
	v, err := scanVendor(r.pool.QueryRow(ctx,
		`SELECT `+vendorColumns+` FROM vendor WHERE id = $1`, id))
	return v, mapError(err)
}

var vendorSortColumns = map[string]string{"id": "id", "nama": "nama", "created_at": "created_at"}

func (r *vendorRepository) FindAll(ctx context.Context, q model.ListQuery) ([]model.Vendor, int, error) {
	where, args := "WHERE 1=1", []any{}
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND (nama ILIKE $%d OR telepon ILIKE $%d)", len(args), len(args))
	}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		where += fmt.Sprintf(" AND is_active = $%d", len(args))
	}

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM vendor "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	sql := fmt.Sprintf("SELECT %s FROM vendor %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		vendorColumns, where, sortColumn(vendorSortColumns, q.Sort), orderDir(q.Order),
		len(args)-1, len(args))
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []model.Vendor{}
	for rows.Next() {
		v, err := scanVendor(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, v)
	}
	return items, total, rows.Err()
}

func (r *vendorRepository) Create(ctx context.Context, v model.Vendor) (model.Vendor, error) {
	out, err := scanVendor(r.pool.QueryRow(ctx,
		`INSERT INTO vendor (nama, alamat, telepon) VALUES ($1, $2, $3) RETURNING `+vendorColumns,
		v.Nama, v.Alamat, v.Telepon))
	return out, mapError(err)
}

func (r *vendorRepository) Update(ctx context.Context, v model.Vendor) (model.Vendor, error) {
	out, err := scanVendor(r.pool.QueryRow(ctx,
		`UPDATE vendor SET nama = $1, alamat = $2, telepon = $3 WHERE id = $4 RETURNING `+vendorColumns,
		v.Nama, v.Alamat, v.Telepon, v.ID))
	return out, mapError(err)
}

// Deactivate = soft delete
func (r *vendorRepository) Deactivate(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, `UPDATE vendor SET is_active = FALSE WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}