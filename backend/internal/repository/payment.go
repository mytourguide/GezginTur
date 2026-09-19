package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"seyahat/backend/internal/models"
)

type PaymentRepository struct{ pool *pgxpool.Pool }

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository { return &PaymentRepository{pool} }

// QueryRow, saglik kontrolleri gibi tekil raw SQL okumalarinda kullanilir.
func (r *PaymentRepository) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return r.pool.QueryRow(ctx, sql, args...)
}


func (r *PaymentRepository) Create(ctx context.Context, p *models.Payment) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO payments (booking_id, iyzico_payment_id, amount, status, installment)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		p.BookingID, p.IyzicoPaymentID, p.Amount, p.Status, p.Installment).
		Scan(&p.ID, &p.CreatedAt)
}

func (r *PaymentRepository) GetByBookingID(ctx context.Context, bookingID string) (*models.Payment, error) {
	var p models.Payment
	err := r.pool.QueryRow(ctx, `
		SELECT id, booking_id, COALESCE(iyzico_payment_id,''), amount, status, installment, created_at
		FROM payments WHERE booking_id=$1 ORDER BY created_at DESC LIMIT 1`, bookingID).
		Scan(&p.ID, &p.BookingID, &p.IyzicoPaymentID, &p.Amount, &p.Status, &p.Installment, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) GetByIyzicoPaymentID(ctx context.Context, pid string) (*models.Payment, error) {
	var p models.Payment
	err := r.pool.QueryRow(ctx, `
		SELECT id, booking_id, iyzico_payment_id, amount, status, installment, created_at
		FROM payments WHERE iyzico_payment_id=$1`, pid).
		Scan(&p.ID, &p.BookingID, &p.IyzicoPaymentID, &p.Amount, &p.Status, &p.Installment, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdateStatus, odeme durumunu ve iyzico referansini gunceller.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, status, iyzicoID string, installment int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE payments SET status=$1, iyzico_payment_id=$2, installment=$3, updated_at=now() WHERE id=$4`,
		status, iyzicoID, installment, id)
	return err
}

// UpdateStatusByIyzicoID, webhook senaryosunda iyzico payment id uzerinden gunceller.
func (r *PaymentRepository) UpdateStatusByIyzicoID(ctx context.Context, iyzicoID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE payments SET status=$1, updated_at=now() WHERE iyzico_payment_id=$2`, status, iyzicoID)
	return err
}

// MarkRefunded, iade sonrasi odeme durumunu gunceller.
func (r *PaymentRepository) MarkRefunded(ctx context.Context, iyzicoID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE payments SET status=$1, updated_at=now() WHERE iyzico_payment_id=$2`, status, iyzicoID)
	return err
}

func (r *PaymentRepository) List(ctx context.Context) ([]models.Payment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.booking_id, COALESCE(p.iyzico_payment_id,''), p.amount, p.status, p.installment, p.created_at
		FROM payments p ORDER BY p.created_at DESC LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Payment, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var p models.Payment
		if err := rows.Scan(&p.ID, &p.BookingID, &p.IyzicoPaymentID, &p.Amount, &p.Status, &p.Installment, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
