package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"seyahat/backend/internal/models"
)

type CouponRepository struct{ pool *pgxpool.Pool }

func NewCouponRepository(pool *pgxpool.Pool) *CouponRepository { return &CouponRepository{pool} }

// GetByCode, aktif ve tarih araligi gecerli kuponu dondurur (rezervasyonda indirim icin).
func (r *CouponRepository) GetByCode(ctx context.Context, code string) (*models.Coupon, error) {
	var c models.Coupon
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, discount_type, discount_value,
		       COALESCE(to_char(valid_from,'YYYY-MM-DD'),''),
		       COALESCE(to_char(valid_until,'YYYY-MM-DD'),''),
		       active, created_at
		FROM coupons
		WHERE code = $1 AND active = true
		  AND (valid_from IS NULL OR valid_from <= CURRENT_DATE)
		  AND (valid_until IS NULL OR valid_until >= CURRENT_DATE)`, code).
		Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.ValidFrom, &c.ValidUntil, &c.Active, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CouponRepository) List(ctx context.Context) ([]models.Coupon, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, discount_type, discount_value,
		       COALESCE(to_char(valid_from,'YYYY-MM-DD'),''),
		       COALESCE(to_char(valid_until,'YYYY-MM-DD'),''),
		       active, created_at
		FROM coupons ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Coupon, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var c models.Coupon
		if err := rows.Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue,
			&c.ValidFrom, &c.ValidUntil, &c.Active, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CouponRepository) Create(ctx context.Context, code, discountType string, value float64, validFrom, validUntil string) (*models.Coupon, error) {
	var c models.Coupon
	err := r.pool.QueryRow(ctx, `
		INSERT INTO coupons (code, discount_type, discount_value, valid_from, valid_until)
		VALUES ($1,$2,$3, NULLIF($4,'')::date, NULLIF($5,'')::date)
		RETURNING id, code, discount_type, discount_value,
		          COALESCE(to_char(valid_from,'YYYY-MM-DD'),''),
		          COALESCE(to_char(valid_until,'YYYY-MM-DD'),''),
		          active, created_at`,
		code, discountType, value, validFrom, validUntil).
		Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.ValidFrom, &c.ValidUntil, &c.Active, &c.CreatedAt)
	return &c, err
}

func (r *CouponRepository) SetActive(ctx context.Context, id string, active bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE coupons SET active=$1 WHERE id=$2`, active, id)
	return err
}

func (r *CouponRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM coupons WHERE id=$1`, id)
	return err
}
