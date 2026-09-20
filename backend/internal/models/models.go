package models

import "time"

// --- Kullanici ---
type User struct {
	ID           string    `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	PasswordHash string    `json:"-"` // asla JSON'a cikmasin
	TotpSecret   string    `json:"-"`
	Role         string    `json:"role"`
	TotpEnabled  bool      `json:"totp_enabled"`
	CreatedAt    time.Time `json:"created_at"`
}

// LangEN: istemcinin "en" dil isteginde ceviri alanlari ana alanlara kopyalanir.
// (bos ceviri varsa TR icerik oldugu gibi kalir)
func (t *Tour) Localize() {
	if t.TitleEN != "" {
		t.Title = t.TitleEN
	}
	if t.SlugEN != "" {
		t.Slug = t.SlugEN
	}
	if t.DescriptionEN != "" {
		t.Description = t.DescriptionEN
	}
	if t.LocationEN != "" {
		t.Location = t.LocationEN
	}
	if t.CategoryNameEN != "" {
		t.CategoryName = t.CategoryNameEN
	}
	for i := range t.Included {
		if t.Included[i].DescriptionEN != "" {
			t.Included[i].Description = t.Included[i].DescriptionEN
		}
	}
	for i := range t.Excluded {
		if t.Excluded[i].DescriptionEN != "" {
			t.Excluded[i].Description = t.Excluded[i].DescriptionEN
		}
	}
	for i := range t.BringItems {
		if t.BringItems[i].DescriptionEN != "" {
			t.BringItems[i].Description = t.BringItems[i].DescriptionEN
		}
	}
	for i := range t.Itinerary {
		if t.Itinerary[i].TitleEN != "" {
			t.Itinerary[i].Title = t.Itinerary[i].TitleEN
		}
		if t.Itinerary[i].DescriptionEN != "" {
			t.Itinerary[i].Description = t.Itinerary[i].DescriptionEN
		}
	}
}

// Localize: kategori icin EN ceviri uygulama
func (c *Category) Localize() {
	if c.NameEN != "" {
		c.Name = c.NameEN
	}
	if c.SlugEN != "" {
		c.Slug = c.SlugEN
	}
}

// --- Tur kategorisi ---
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`

	// Dil destegi: en cevirisi (JSON'a cikmaz; API katmani uygular)
	NameEN string `json:"-"`
	SlugEN string `json:"-"`
}

// --- Tur ---
type Tour struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	CategoryID     string    `json:"category_id"`
	CategoryName   string    `json:"category_name,omitempty"`
	CoverImage     string    `json:"cover_image"`
	DurationDays   int       `json:"duration_days"`
	DurationNights int       `json:"duration_nights"`
	Location       string    `json:"location"`
	BasePrice      float64   `json:"base_price"`
	Currency       string    `json:"currency"`
	Active         bool      `json:"active"`
	Featured       bool      `json:"featured"`       // anasayfada yayinla
	PublishOrder   int       `json:"publish_order"`  // anasayfa siralamasi (buyuk once)
	CreatedAt      time.Time `json:"created_at"`

	// Dil destegi: en cevirileri (JSON'a cikmaz; API katmani uygular)
	TitleEN        string `json:"-"`
	SlugEN         string `json:"-"`
	DescriptionEN  string `json:"-"`
	LocationEN     string `json:"-"`
	CategoryNameEN string `json:"-"`

	Images     []TourImage    `json:"images,omitempty"`
	Departures []Departure    `json:"departures,omitempty"`
	Included   []TourItem     `json:"included,omitempty"`
	Excluded   []TourItem     `json:"excluded,omitempty"`
	BringItems []TourItem     `json:"bring_items,omitempty"`
	Itinerary  []ItineraryDay `json:"itinerary,omitempty"`
}

type TourImage struct {
	ID        string `json:"id"`
	TourID    string `json:"tour_id"`
	ImageURL  string `json:"image_url"`
	SortOrder int    `json:"sort_order"`
}

// --- Kalkis (tarih bazli fiyat + kontenjan) ---
type Departure struct {
	ID        string    `json:"id"`
	TourID    string    `json:"tour_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Capacity  int       `json:"capacity"`
	Filled    int       `json:"filled"`
	Price     float64   `json:"price"`
	Available int       `json:"available,omitempty"` // hesaplanmis alan
}

// --- Dahil / Haric / Yaninda getirilecekler / Program ---
type TourItem struct {
	ID          string `json:"id"`
	TourID      string `json:"tour_id"`
	Description string `json:"description"`

	DescriptionEN string `json:"-"` // dil destegi
}

type ItineraryDay struct {
	ID          string `json:"id"`
	TourID      string `json:"tour_id"`
	DayNo       int    `json:"day_no"`
	Title       string `json:"title"`
	Description string `json:"description"`

	TitleEN       string `json:"-"` // dil destegi
	DescriptionEN string `json:"-"` // dil destegi
}

// --- Rezervasyon ---
type Booking struct {
	ID          string    `json:"id"`
	TourID      string    `json:"tour_id"`
	TourTitle   string    `json:"tour_title,omitempty"`
	DepartureID string    `json:"departure_id"`
	UserID      string    `json:"user_id"`
	AdultCount  int       `json:"adult_count"`
	ChildCount  int       `json:"child_count"`
	TotalPrice  float64   `json:"total_price"`
	Status      string    `json:"status"` // pending | paid | confirmed | cancelled | failed
	CreatedAt   time.Time `json:"created_at"`

	CouponCode     string  `json:"coupon_code,omitempty"`
	DiscountAmount float64 `json:"discount_amount,omitempty"`

	Travelers []Traveler `json:"travelers,omitempty"`
}

// --- Kupon ---
type Coupon struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	DiscountType  string    `json:"discount_type"` // percent | fixed
	DiscountValue float64   `json:"discount_value"`
	ValidFrom     string    `json:"valid_from,omitempty"`  // YYYY-MM-DD (bos = sinirsiz)
	ValidUntil    string    `json:"valid_until,omitempty"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
}

type Traveler struct {
	ID        string    `json:"id"`
	BookingID string    `json:"booking_id"`
	FullName  string    `json:"full_name"`
	BirthDate time.Time `json:"birth_date"`
	IDNumber  string    `json:"id_number,omitempty"` // yalnizca maskele son 4 hane doner
	IDLast4   string    `json:"id_last4,omitempty"`
	IsChild   bool      `json:"is_child"`
}

// --- Odeme ---
type Payment struct {
	ID              string    `json:"id"`
	BookingID       string    `json:"booking_id"`
	IyzicoPaymentID string    `json:"iyzico_payment_id"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"` // initiated | success | failed | refunded | partial_refunded
	Installment     int       `json:"installment"`
	CardType        string    `json:"card_type,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// --- Yardimci istek tipleri ---
type CreateBookingRequest struct {
	DepartureID string `json:"departure_id"`
	CouponCode  string `json:"coupon_code"` // opsiyonel indirim kodu
	AdultCount  int    `json:"adult_count"`
	ChildCount  int    `json:"child_count"`
	Travelers   []struct {
		FullName  string `json:"full_name"`
		BirthDate string `json:"birth_date"` // YYYY-MM-DD
		IDNumber  string `json:"id_number"`
	} `json:"travelers"`
}

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TotpCode string `json:"totp_code"` // 2FA aciksa zorunlu
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// --- Admin: tur olusturma/guncelleme (ic ice listeler tek istekte) ---
type AdminTourInput struct {
	Title          string         `json:"title"`
	Slug           string         `json:"slug"`
	Description    string         `json:"description"`
	CategoryID     string         `json:"category_id"`
	CoverImage     string         `json:"cover_image"`
	DurationDays   int            `json:"duration_days"`
	DurationNights int            `json:"duration_nights"`
	Location       string         `json:"location"`
	BasePrice      float64        `json:"base_price"`
	Currency       string         `json:"currency"`
	Active         bool           `json:"active"`
	Featured       bool           `json:"featured"`
	PublishOrder   int            `json:"publish_order"`
	Images         []string       `json:"images"`
	Included       []string       `json:"included"`
	Excluded       []string       `json:"excluded"`
	BringItems     []string       `json:"bring_items"`
	Itinerary      []ItineraryDay `json:"itinerary"`
}
