package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"seyahat/backend/internal/models"
)

type BookingRepository struct{ pool *pgxpool.Pool }

func NewBookingRepository(pool *pgxpool.Pool) *BookingRepository { return &BookingRepository{pool} }

// QueryRow, saglik kontrolleri gibi tekil raw SQL okumalarinda kullanilir.
func (r *BookingRepository) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return r.pool.QueryRow(ctx, sql, args...)
}


// BeginTx, cok adimli rezervasyon islemleri icin transaction baslatir.
func (r *BookingRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// CreateWithTx, rezervasyonu + yolculari verilen transaction icinde olusturur.
// Yolcu kimlik/pasaport numaralari (id_number_encrypted) servis katmaninda sifrelenmis
// olarak gelir; duz metin hicbir zaman veritabanina yazilmaz.
func (r *BookingRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, b *models.Booking, travelers []models.Traveler) error {
	err := tx.QueryRow(ctx, `
		INSERT INTO bookings (tour_id, departure_id, user_id, adult_count, child_count, total_price, status, coupon_code, discount_amount)
		VALUES ($1,$2,$3,$4,$5,$6,'pending', NULLIF($7,''), $8) RETURNING id, created_at`,
		b.TourID, b.DepartureID, b.UserID, b.AdultCount, b.ChildCount, b.TotalPrice, b.CouponCode, b.DiscountAmount).
		Scan(&b.ID, &b.CreatedAt)
	if err != nil {
		return err
	}
	for i := range travelers {
		tr := &travelers[i]
		if err := tx.QueryRow(ctx, `
			INSERT INTO booking_travelers (booking_id, full_name, birth_date, id_number_encrypted, id_number_last4, is_child)
			VALUES ($1,$2,$3::date,$4,$5,$6) RETURNING id`,
			b.ID, tr.FullName, tr.BirthDate.Format("2006-01-02"), tr.IDNumber, tr.IDLast4, tr.IsChild).
			Scan(&tr.ID); err != nil {
			return err
		}
	}
	return nil
}

// FillCapacity, kalkisin dolu kontenjanini artirir (FOR UPDATE kilidiyle birlikte cagrilir).
func (r *BookingRepository) FillCapacity(ctx context.Context, tx pgx.Tx, departureID string, count int) error {
	_, err := tx.Exec(ctx, `UPDATE tour_departures SET filled = filled + $1 WHERE id=$2`, count, departureID)
	return err
}

// RefundCapacity, iptal/iade durumunda kontenjani geri verir.
func (r *BookingRepository) RefundCapacity(ctx context.Context, tx pgx.Tx, departureID string, count int) error {
	_, err := tx.Exec(ctx, `UPDATE tour_departures SET filled = GREATEST(filled - $1, 0) WHERE id=$2`, count, departureID)
	return err
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*models.Booking, error) {
	var b models.Booking
	err := r.pool.QueryRow(ctx, `
		SELECT b.id, b.tour_id, t.title, b.departure_id, b.user_id, b.adult_count, b.child_count,
		       b.total_price, b.status, b.created_at
		FROM bookings b JOIN tours t ON t.id = b.tour_id WHERE b.id = $1`, id).
		Scan(&b.ID, &b.TourID, &b.TourTitle, &b.DepartureID, &b.UserID, &b.AdultCount,
			&b.ChildCount, &b.TotalPrice, &b.Status, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b.Travelers = r.travelers(ctx, id)
	return &b, nil
}

// travelers: KVKK maskeleme — kimlik numarasi yerine yalnizca son 4 hane doner.
func (r *BookingRepository) travelers(ctx context.Context, bookingID string) []models.Traveler {
	rows, err := r.pool.Query(ctx, `
		SELECT id, booking_id, full_name, birth_date, COALESCE(id_number_last4,''), is_child
		FROM booking_travelers WHERE booking_id=$1 ORDER BY birth_date DESC`, bookingID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	out := make([]models.Traveler, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var tr models.Traveler
		if err := rows.Scan(&tr.ID, &tr.BookingID, &tr.FullName, &tr.BirthDate, &tr.IDLast4, &tr.IsChild); err == nil {
			out = append(out, tr)
		}
	}
	return out
}

func (r *BookingRepository) ListByUser(ctx context.Context, userID string) ([]models.Booking, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT b.id, b.tour_id, t.title, b.departure_id, b.user_id,
		       b.adult_count, b.child_count, b.total_price, b.status, b.created_at
		FROM bookings b JOIN tours t ON t.id = b.tour_id
		WHERE b.user_id = $1 ORDER BY b.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Booking, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.TourID, &b.TourTitle, &b.DepartureID, &b.UserID,
			&b.AdultCount, &b.ChildCount, &b.TotalPrice, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// BookingFilters, admin listeleme filtreleri.
type BookingFilters struct {
	Status   string
	TourID   string
	FromDate string
	ToDate   string
	Search   string // musteri adi / e-posta
}

func (r *BookingRepository) List(ctx context.Context, f BookingFilters) ([]models.Booking, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	i := 1
	if f.Status != "" {
		where = append(where, fmt.Sprintf("b.status=$%d", i))
		args = append(args, f.Status)
		i++
	}
	if f.TourID != "" {
		where = append(where, fmt.Sprintf("b.tour_id=$%d", i))
		args = append(args, f.TourID)
		i++
	}
	if f.FromDate != "" {
		where = append(where, fmt.Sprintf("b.created_at::date >= $%d::date", i))
		args = append(args, f.FromDate)
		i++
	}
	if f.ToDate != "" {
		where = append(where, fmt.Sprintf("b.created_at::date <= $%d::date", i))
		args = append(args, f.ToDate)
		i++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(u.full_name ILIKE '%%'||$%d||'%%' OR u.email ILIKE '%%'||$%d||'%%')", i, i))
		args = append(args, f.Search)
		i++
	}
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT b.id, b.tour_id, t.title, b.departure_id, b.user_id,
		       b.adult_count, b.child_count, b.total_price, b.status, b.created_at
		FROM bookings b
		JOIN tours t ON t.id = b.tour_id
		JOIN users u ON u.id = b.user_id
		WHERE %s ORDER BY b.created_at DESC LIMIT 500`, strings.Join(where, " AND ")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Booking, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.TourID, &b.TourTitle, &b.DepartureID, &b.UserID,
			&b.AdultCount, &b.ChildCount, &b.TotalPrice, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// UpdateStatus, rezervasyon durumunu gunceller (pending → paid → confirmed → cancelled/failed).
func (r *BookingRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE bookings SET status=$1 WHERE id=$2`, status, id)
	return err
}

// --- Admin dashboard istatistikleri ---

type DashboardStats struct {
	TotalBookings   int     `json:"total_bookings"`
	PendingPayments int     `json:"pending_payments"`
	TotalRevenue    float64 `json:"total_revenue"`
	TotalCustomers  int     `json:"total_customers"`
}

func (r *BookingRepository) DashboardStats(ctx context.Context) (*DashboardStats, error) {
	var s DashboardStats
	err := r.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM bookings),
			(SELECT COUNT(*) FROM bookings WHERE status='pending'),
			(SELECT COALESCE(SUM(total_price),0) FROM bookings WHERE status IN ('paid','confirmed')),
			(SELECT COUNT(*) FROM users WHERE role='customer')`).Scan(
		&s.TotalBookings, &s.PendingPayments, &s.TotalRevenue, &s.TotalCustomers)
	return &s, err
}

// SalesPoint, satis grafigi icin gunluk toplam.
type SalesPoint struct {
	Date  time.Time `json:"date"`
	Total float64   `json:"total"`
}

// SalesSeries, son N gunun gunluk satis toplamini dondurur.
func (r *BookingRepository) SalesSeries(ctx context.Context, days int) ([]SalesPoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT d::date, COALESCE(SUM(b.total_price),0)
		FROM generate_series(CURRENT_DATE - ($1 * interval '1 day'), CURRENT_DATE, '1 day') d
		LEFT JOIN bookings b ON b.created_at::date = d::date AND b.status IN ('paid','confirmed')
		GROUP BY d::date ORDER BY d::date`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SalesPoint, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var p SalesPoint
		if err := rows.Scan(&p.Date, &p.Total); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// OccupancyRow, kalkis bazli doluluk bilgisi.
type OccupancyRow struct {
	Tour      string    `json:"tour"`
	StartDate time.Time `json:"start_date"`
	Capacity  int       `json:"capacity"`
	Filled    int       `json:"filled"`
}

func (r *BookingRepository) Occupancy(ctx context.Context) ([]OccupancyRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.title, d.start_date, d.capacity, d.filled
		FROM tour_departures d JOIN tours t ON t.id = d.tour_id
		WHERE d.start_date >= CURRENT_DATE ORDER BY d.start_date LIMIT 20`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]OccupancyRow, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var row OccupancyRow
		if err := rows.Scan(&row.Tour, &row.StartDate, &row.Capacity, &row.Filled); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// Cancel, bekleyen (odemesi alinmamis) rezervasyonu iptal eder ve kontenjani iade eder.
func (r *BookingRepository) Cancel(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var depID, status string
	var adults, children int
	err = tx.QueryRow(ctx, `SELECT departure_id, adult_count, child_count, status FROM bookings WHERE id=$1 FOR UPDATE`, id).
		Scan(&depID, &adults, &children, &status)
	if err != nil {
		return err
	}
	if status == "paid" || status == "confirmed" {
		return errors.New("odemesi alinmis rezervasyonun iptali admin panelinden iade ile yapilmalidir")
	}
	if _, err := tx.Exec(ctx, `UPDATE bookings SET status='cancelled' WHERE id=$1`, id); err != nil {
		return err
	}
	if err := r.RefundCapacity(ctx, tx, depID, adults+children); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
