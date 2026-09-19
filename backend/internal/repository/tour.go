package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"seyahat/backend/internal/models"
)

type TourRepository struct{ pool *pgxpool.Pool }

func NewTourRepository(pool *pgxpool.Pool) *TourRepository { return &TourRepository{pool} }

// QueryRow, saglik kontrolleri gibi tekil raw SQL okumalarinda kullanilir.
func (r *TourRepository) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return r.pool.QueryRow(ctx, sql, args...)
}


// TourFilters, musteri tarafindaki liste filtrelerini tutar.
type TourFilters struct {
	CategorySlug string
	Search       string // baslik / lokasyon aramasi
	MinPrice     *float64
	MaxPrice     *float64
	DurationMax  *int   // maksimum gun
	StartDate    string // YYYY-MM-DD
	EndDate      string
	Sort         string // price_asc | price_desc | popular | date
	Featured     bool
	Page, Limit  int
}

// List, filtreleri uygulayarak turlari listeler (musteri tarafi, yalniz aktif turlar).
func (r *TourRepository) List(ctx context.Context, f TourFilters) ([]models.Tour, int, error) {
	where := []string{"t.active = true"}
	args := []interface{}{}
	i := 1

	if f.CategorySlug != "" {
		where = append(where, fmt.Sprintf("(c.slug = $%d OR c.slug_en = $%d)", i, i))
		args = append(args, f.CategorySlug)
		i++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(t.title ILIKE '%%'||$%d||'%%' OR t.location ILIKE '%%'||$%d||'%%')", i, i))
		args = append(args, f.Search)
		i++
	}
	if f.MinPrice != nil {
		where = append(where, fmt.Sprintf("t.base_price >= $%d", i))
		args = append(args, *f.MinPrice)
		i++
	}
	if f.MaxPrice != nil {
		where = append(where, fmt.Sprintf("t.base_price <= $%d", i))
		args = append(args, *f.MaxPrice)
		i++
	}
	if f.DurationMax != nil {
		where = append(where, fmt.Sprintf("t.duration_days <= $%d", i))
		args = append(args, *f.DurationMax)
		i++
	}
	if f.Featured {
		where = append(where, "t.featured = true")
	}
	if f.StartDate != "" && f.EndDate != "" {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM tour_departures d WHERE d.tour_id = t.id
			AND d.start_date >= $%d::date AND d.start_date <= $%d::date)`, i, i+1))
		args = append(args, f.StartDate, f.EndDate)
		i += 2
	}

	orderBy := "t.created_at DESC"
	switch f.Sort {
	case "price_asc":
		orderBy = "t.base_price ASC"
	case "price_desc":
		orderBy = "t.base_price DESC"
	case "popular":
		orderBy = "COALESCE(bk.cnt, 0) DESC"
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 12
	}
	offset := (f.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM tours t JOIN tour_categories c ON c.id = t.category_id
		LEFT JOIN (SELECT tour_id, COUNT(*) cnt FROM bookings GROUP BY tour_id) bk ON bk.tour_id = t.id
		WHERE %s`, whereSQL), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT t.id, t.title, t.slug, t.description, t.category_id, c.name, t.cover_image,
		       t.duration_days, t.duration_nights, t.location, t.base_price, t.currency,
		       t.active, t.featured, t.created_at, COALESCE(bk.cnt,0),
		       COALESCE(t.title_en,''), COALESCE(t.slug_en,''),
		       COALESCE(t.description_en,''), COALESCE(t.location_en,''),
		       COALESCE(c.name_en,'')
		FROM tours t
		JOIN tour_categories c ON c.id = t.category_id
		LEFT JOIN (SELECT tour_id, COUNT(*) cnt FROM bookings GROUP BY tour_id) bk ON bk.tour_id = t.id
		WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d`, whereSQL, orderBy, i, i+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]models.Tour, 0) // bos olsa bile JSON'da null degil [] doner
	for rows.Next() {
		var t models.Tour
		var cnt int
		if err := rows.Scan(&t.ID, &t.Title, &t.Slug, &t.Description, &t.CategoryID, &t.CategoryName,
			&t.CoverImage, &t.DurationDays, &t.DurationNights, &t.Location, &t.BasePrice, &t.Currency,
			&t.Active, &t.Featured, &t.CreatedAt, &cnt,
			&t.TitleEN, &t.SlugEN, &t.DescriptionEN, &t.LocationEN, &t.CategoryNameEN); err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

// GetBySlug, tur detayini tum iliskili listeleriyle dondurur.
func (r *TourRepository) GetBySlug(ctx context.Context, slug string) (*models.Tour, error) {
	var t models.Tour
	err := r.pool.QueryRow(ctx, `
		SELECT t.id, t.title, t.slug, t.description, t.category_id, c.name, t.cover_image,
		       t.duration_days, t.duration_nights, t.location, t.base_price, t.currency,
		       t.active, t.featured, t.created_at,
		       COALESCE(t.title_en,''), COALESCE(t.slug_en,''),
		       COALESCE(t.description_en,''), COALESCE(t.location_en,''),
		       COALESCE(c.name_en,'')
		FROM tours t JOIN tour_categories c ON c.id = t.category_id
		WHERE (t.slug = $1 OR t.slug_en = $1) AND t.active = true`, slug).
		Scan(&t.ID, &t.Title, &t.Slug, &t.Description, &t.CategoryID, &t.CategoryName,
			&t.CoverImage, &t.DurationDays, &t.DurationNights, &t.Location, &t.BasePrice,
			&t.Currency, &t.Active, &t.Featured, &t.CreatedAt,
			&t.TitleEN, &t.SlugEN, &t.DescriptionEN, &t.LocationEN, &t.CategoryNameEN)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	t.Images, _ = r.images(ctx, t.ID)
	t.Departures, _ = r.departures(ctx, t.ID, true) // yalnizca ileri tarihli, bos yeri olanlar
	t.Included, _ = r.items(ctx, "tour_included_items", t.ID)
	t.Excluded, _ = r.items(ctx, "tour_excluded_items", t.ID)
	t.BringItems, _ = r.items(ctx, "tour_bring_items", t.ID)

	rows, err := r.pool.Query(ctx, `
		SELECT id, tour_id, day_no, title, description, COALESCE(title_en,''), COALESCE(description_en,'')
		FROM tour_itinerary
		WHERE tour_id=$1 ORDER BY day_no ASC`, t.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var d models.ItineraryDay
			if rows.Scan(&d.ID, &d.TourID, &d.DayNo, &d.Title, &d.Description, &d.TitleEN, &d.DescriptionEN) == nil {
				t.Itinerary = append(t.Itinerary, d)
			}
		}
	}
	return &t, nil
}

func (r *TourRepository) images(ctx context.Context, tourID string) ([]models.TourImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tour_id, image_url, sort_order FROM tour_images
		WHERE tour_id=$1 ORDER BY sort_order ASC`, tourID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.TourImage, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var img models.TourImage
		if err := rows.Scan(&img.ID, &img.TourID, &img.ImageURL, &img.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

// departures: onlyFuture=true ise gecmis tarihli ve dolu kalkislari dislar (musteri tarafi).
func (r *TourRepository) departures(ctx context.Context, tourID string, onlyFuture bool) ([]models.Departure, error) {
	q := `SELECT id, tour_id, start_date, end_date, capacity, filled, price
	      FROM tour_departures WHERE tour_id=$1`
	if onlyFuture {
		q += ` AND start_date >= CURRENT_DATE AND filled < capacity ORDER BY start_date ASC`
	} else {
		q += ` ORDER BY start_date DESC`
	}
	rows, err := r.pool.Query(ctx, q, tourID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Departure, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var d models.Departure
		if err := rows.Scan(&d.ID, &d.TourID, &d.StartDate, &d.EndDate, &d.Capacity, &d.Filled, &d.Price); err != nil {
			return nil, err
		}
		d.Available = d.Capacity - d.Filled
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *TourRepository) items(ctx context.Context, table, tourID string) ([]models.TourItem, error) {
	rows, err := r.pool.Query(ctx, fmt.Sprintf(
		`SELECT id, tour_id, description, COALESCE(description_en,'') FROM %s WHERE tour_id=$1 ORDER BY id ASC`, table), tourID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.TourItem, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var it models.TourItem
		if err := rows.Scan(&it.ID, &it.TourID, &it.Description, &it.DescriptionEN); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// GetDepartureByIDForUpdate, rezervasyon sirasinda kontenjan yarisini onlemek icin
// satir kilitli (FOR UPDATE) okur.
func (r *TourRepository) GetDepartureByIDForUpdate(ctx context.Context, tx pgx.Tx, id string) (*models.Departure, error) {
	var d models.Departure
	err := tx.QueryRow(ctx, `
		SELECT id, tour_id, start_date, end_date, capacity, filled, price
		FROM tour_departures WHERE id=$1 FOR UPDATE`, id).
		Scan(&d.ID, &d.TourID, &d.StartDate, &d.EndDate, &d.Capacity, &d.Filled, &d.Price)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// --- Kategoriler ---

func (r *TourRepository) Categories(ctx context.Context) ([]models.Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, slug, COALESCE(description,''), COALESCE(name_en,''), COALESCE(slug_en,'') FROM tour_categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Category, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.NameEN, &c.SlugEN); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *TourRepository) CreateCategory(ctx context.Context, name, slug, description string) (*models.Category, error) {
	var c models.Category
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tour_categories (name, slug, description) VALUES ($1,$2,$3)
		RETURNING id, name, slug, COALESCE(description,'')`, name, slug, description).
		Scan(&c.ID, &c.Name, &c.Slug, &c.Description)
	return &c, err
}

func (r *TourRepository) UpdateCategory(ctx context.Context, id, name, slug, description string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tour_categories SET name=$1, slug=$2, description=$3 WHERE id=$4`,
		name, slug, description, id)
	return err
}

// --- Admin: tur CRUD ---

func (r *TourRepository) ListAdmin(ctx context.Context, search string) ([]models.Tour, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.title, t.slug, t.description, t.category_id, c.name, t.cover_image,
		       t.duration_days, t.duration_nights, t.location, t.base_price, t.currency,
		       t.active, t.featured, t.created_at
		FROM tours t JOIN tour_categories c ON c.id = t.category_id
		WHERE ($1 = '' OR t.title ILIKE '%'||$1||'%')
		ORDER BY t.created_at DESC`, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Tour, 0) // bos sonucta JSON null degil [] doner
	for rows.Next() {
		var t models.Tour
		if err := rows.Scan(&t.ID, &t.Title, &t.Slug, &t.Description, &t.CategoryID, &t.CategoryName,
			&t.CoverImage, &t.DurationDays, &t.DurationNights, &t.Location, &t.BasePrice, &t.Currency,
			&t.Active, &t.Featured, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TourRepository) GetByID(ctx context.Context, id string) (*models.Tour, error) {
	var t models.Tour
	err := r.pool.QueryRow(ctx, `
		SELECT t.id, t.title, t.slug, t.description, t.category_id, c.name, t.cover_image,
		       t.duration_days, t.duration_nights, t.location, t.base_price, t.currency,
		       t.active, t.featured, t.created_at
		FROM tours t JOIN tour_categories c ON c.id = t.category_id WHERE t.id = $1`, id).
		Scan(&t.ID, &t.Title, &t.Slug, &t.Description, &t.CategoryID, &t.CategoryName,
			&t.CoverImage, &t.DurationDays, &t.DurationNights, &t.Location, &t.BasePrice,
			&t.Currency, &t.Active, &t.Featured, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.Images, _ = r.images(ctx, t.ID)
	t.Departures, _ = r.departures(ctx, t.ID, false)
	t.Included, _ = r.items(ctx, "tour_included_items", t.ID)
	t.Excluded, _ = r.items(ctx, "tour_excluded_items", t.ID)
	t.BringItems, _ = r.items(ctx, "tour_bring_items", t.ID)
	rows, _ := r.pool.Query(ctx, `SELECT id, tour_id, day_no, title, description FROM tour_itinerary WHERE tour_id=$1 ORDER BY day_no`, t.ID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var d models.ItineraryDay
			if rows.Scan(&d.ID, &d.TourID, &d.DayNo, &d.Title, &d.Description) == nil {
				t.Itinerary = append(t.Itinerary, d)
			}
		}
	}
	return &t, nil
}

// Create, turu ve ic ice listelerini tek transaction'da olusturur.
func (r *TourRepository) Create(ctx context.Context, in *models.AdminTourInput) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO tours (title, slug, description, category_id, cover_image, duration_days,
			duration_nights, location, base_price, currency, active, featured)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id`,
		in.Title, in.Slug, in.Description, in.CategoryID, in.CoverImage, in.DurationDays,
		in.DurationNights, in.Location, in.BasePrice, in.Currency, in.Active, in.Featured).Scan(&id)
	if err != nil {
		return "", err
	}

	if err := r.replaceChildren(ctx, tx, id, in); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return id, nil
}

// Update, turu ve ic ice listelerini tek transaction'da gunceller.
func (r *TourRepository) Update(ctx context.Context, id string, in *models.AdminTourInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE tours SET title=$1, slug=$2, description=$3, category_id=$4, cover_image=$5,
			duration_days=$6, duration_nights=$7, location=$8, base_price=$9, currency=$10,
			active=$11, featured=$12, updated_at=now() WHERE id=$13`,
		in.Title, in.Slug, in.Description, in.CategoryID, in.CoverImage, in.DurationDays,
		in.DurationNights, in.Location, in.BasePrice, in.Currency, in.Active, in.Featured, id)
	if err != nil {
		return err
	}

	for _, table := range []string{"tour_images", "tour_included_items", "tour_excluded_items", "tour_bring_items", "tour_itinerary"} {
		if _, err := tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE tour_id=$1", table), id); err != nil {
			return err
		}
	}
	if err := r.replaceChildren(ctx, tx, id, in); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// replaceChildren, gorseller/dahil/haric/getirilecekler/program kayitlarini yazar.
func (r *TourRepository) replaceChildren(ctx context.Context, tx pgx.Tx, tourID string, in *models.AdminTourInput) error {
	for i, url := range in.Images {
		if _, err := tx.Exec(ctx, `INSERT INTO tour_images (tour_id, image_url, sort_order) VALUES ($1,$2,$3)`,
			tourID, url, i); err != nil {
			return err
		}
	}
	insertItems := func(table string, items []string) error {
		for _, desc := range items {
			if _, err := tx.Exec(ctx, fmt.Sprintf("INSERT INTO %s (tour_id, description) VALUES ($1,$2)", table),
				tourID, desc); err != nil {
				return err
			}
		}
		return nil
	}
	if err := insertItems("tour_included_items", in.Included); err != nil {
		return err
	}
	if err := insertItems("tour_excluded_items", in.Excluded); err != nil {
		return err
	}
	if err := insertItems("tour_bring_items", in.BringItems); err != nil {
		return err
	}
	for _, d := range in.Itinerary {
		if _, err := tx.Exec(ctx, `
			INSERT INTO tour_itinerary (tour_id, day_no, title, description) VALUES ($1,$2,$3,$4)`,
			tourID, d.DayNo, d.Title, d.Description); err != nil {
			return err
		}
	}
	return nil
}

func (r *TourRepository) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tours WHERE id=$1`, id)
	return err
}

// --- Admin: kalkis tarihleri ---

func (r *TourRepository) CreateDeparture(ctx context.Context, tourID, startDate, endDate string, capacity int, price float64) (*models.Departure, error) {
	var d models.Departure
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tour_departures (tour_id, start_date, end_date, capacity, price)
		VALUES ($1,$2::date,$3::date,$4,$5)
		RETURNING id, tour_id, start_date, end_date, capacity, filled, price`,
		tourID, startDate, endDate, capacity, price).
		Scan(&d.ID, &d.TourID, &d.StartDate, &d.EndDate, &d.Capacity, &d.Filled, &d.Price)
	return &d, err
}

func (r *TourRepository) DeleteDeparture(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tour_departures WHERE id=$1 AND filled = 0`, id)
	return err
}
