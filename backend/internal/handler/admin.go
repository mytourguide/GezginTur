package handler

import (
	"net/http"
	"strings"

	"seyahat/backend/internal/models"
	"seyahat/backend/internal/repository"
)

// ---------- Dashboard ----------

// AdminDashboard, satis grafigi + ozet istatistik + doluluk oranlarini dondurur.
func (a *API) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := a.Bookings.DashboardStats(r.Context())
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "istatistikler alinamadi")
		return
	}
	sales, err := a.Bookings.SalesSeries(r.Context(), 30)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "satis verisi alinamadi")
		return
	}
	occupancy, err := a.Bookings.Occupancy(r.Context())
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "doluluk verisi alinamadi")
		return
	}
	JSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats, "sales": sales, "occupancy": occupancy,
	})
}

// ---------- Tur yonetimi ----------

func (a *API) AdminListTours(w http.ResponseWriter, r *http.Request) {
	tours, err := a.Tours.ListAdmin(r.Context(), r.URL.Query().Get("search"))
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "turlar listelenemedi")
		return
	}
	JSON(w, http.StatusOK, tours)
}

func (a *API) AdminGetTour(w http.ResponseWriter, r *http.Request) {
	t, err := a.Tours.GetByID(r.Context(), r.PathValue("id"))
	if err != nil || t == nil {
		ErrorJSON(w, http.StatusNotFound, "tur bulunamadi")
		return
	}
	JSON(w, http.StatusOK, t)
}

// AdminCreateTour, turu tum ic ice listeleriyle birlikte olusturur.
func (a *API) AdminCreateTour(w http.ResponseWriter, r *http.Request) {
	var in models.AdminTourInput
	if err := DecodeBody(r, &in); err != nil || in.Title == "" || in.Slug == "" || in.CategoryID == "" {
		ErrorJSON(w, http.StatusBadRequest, "baslik, slug ve kategori zorunlu")
		return
	}
	if in.Currency == "" {
		in.Currency = "TRY"
	}
	id, err := a.Tours.Create(r.Context(), &in)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "tur olusturulamadi: "+err.Error())
		return
	}
	JSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (a *API) AdminUpdateTour(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var in models.AdminTourInput
	if err := DecodeBody(r, &in); err != nil || in.Title == "" {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz tur verisi")
		return
	}
	if err := a.Tours.Update(r.Context(), id, &in); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "tur guncellenemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) AdminDeleteTour(w http.ResponseWriter, r *http.Request) {
	if err := a.Tours.Delete(r.Context(), r.PathValue("id")); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "tur silinemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------- Kalkis tarihleri ----------

func (a *API) AdminCreateDeparture(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TourID    string  `json:"tour_id"`
		StartDate string  `json:"start_date"`
		EndDate   string  `json:"end_date"`
		Capacity  int     `json:"capacity"`
		Price     float64 `json:"price"`
	}
	if err := DecodeBody(r, &req); err != nil || req.StartDate == "" || req.EndDate == "" || req.Capacity < 1 {
		ErrorJSON(w, http.StatusBadRequest, "tarih ve kontenjan zorunlu")
		return
	}
	d, err := a.Tours.CreateDeparture(r.Context(), req.TourID, req.StartDate, req.EndDate, req.Capacity, req.Price)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kalkis eklenemedi")
		return
	}
	JSON(w, http.StatusCreated, d)
}

func (a *API) AdminDeleteDeparture(w http.ResponseWriter, r *http.Request) {
	if err := a.Tours.DeleteDeparture(r.Context(), r.PathValue("id")); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kalkis silinemedi (satis var olabilir)")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------- Rezervasyon yonetimi ----------

func (a *API) AdminListBookings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := a.Bookings.List(r.Context(), repository.BookingFilters{
		Status: q.Get("status"), TourID: q.Get("tour_id"),
		FromDate: q.Get("from"), ToDate: q.Get("to"), Search: q.Get("search"),
	})
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "rezervasyonlar listelenemedi")
		return
	}
	JSON(w, http.StatusOK, list)
}

func (a *API) AdminGetBooking(w http.ResponseWriter, r *http.Request) {
	b, err := a.Bookings.GetByID(r.Context(), r.PathValue("id"))
	if err != nil || b == nil {
		ErrorJSON(w, http.StatusNotFound, "rezervasyon bulunamadi")
		return
	}
	JSON(w, http.StatusOK, b)
}

// AdminUpdateBookingStatus, rezervasyon durumunu manuel gunceller (orn. paid → confirmed).
func (a *API) AdminUpdateBookingStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Status string `json:"status"`
	}
	if err := DecodeBody(r, &req); err != nil || req.Status == "" {
		ErrorJSON(w, http.StatusBadRequest, "durum zorunlu")
		return
	}
	allowed := map[string]bool{"pending": true, "paid": true, "confirmed": true, "cancelled": true, "failed": true}
	if !allowed[req.Status] {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz durum")
		return
	}
	if err := a.Bookings.UpdateStatus(r.Context(), r.PathValue("id"), req.Status); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "durum guncellenemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// AdminRefund, iyzico uzerinden iade baslatir; basariliysa rezervasyonu iptal eder
// ve kontenjani iade eder.
func (a *API) AdminRefund(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Amount float64 `json:"amount"` // bossa tam iade
	}
	_ = DecodeBody(r, &req)

	b, err := a.Bookings.GetByID(r.Context(), id)
	if err != nil || b == nil {
		ErrorJSON(w, http.StatusNotFound, "rezervasyon bulunamadi")
		return
	}
	p, err := a.Payments.GetByBookingID(r.Context(), id)
	if err != nil || p == nil || p.Status != "success" {
		ErrorJSON(w, http.StatusBadRequest, "iade edilebilir basarili bir odeme yok")
		return
	}
	amount := req.Amount
	if amount <= 0 {
		amount = p.Amount
	}
	if err := a.Payment.Refund(r.Context(), p.IyzicoPaymentID, amount); err != nil {
		ErrorJSON(w, http.StatusBadGateway, "iyzico iade hatasi: "+err.Error())
		return
	}
	newPayStatus := "refunded"
	if amount < p.Amount {
		newPayStatus = "partial_refunded"
	}
	_ = a.Payments.MarkRefunded(r.Context(), p.IyzicoPaymentID, newPayStatus)
	_ = a.Bookings.UpdateStatus(r.Context(), id, "cancelled")
	// Kontenjan iadesi
	tx, err := a.Bookings.BeginTx(r.Context())
	if err == nil {
		_ = a.Bookings.RefundCapacity(r.Context(), tx, b.DepartureID, b.AdultCount+b.ChildCount)
		_ = tx.Commit(r.Context())
	}
	JSON(w, http.StatusOK, map[string]string{"status": "refunded"})
}

// ---------- Odeme kayitlari ----------

func (a *API) AdminListPayments(w http.ResponseWriter, r *http.Request) {
	list, err := a.Payments.List(r.Context())
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "odemeler listelenemedi")
		return
	}
	JSON(w, http.StatusOK, list)
}

// ---------- Musteri yonetimi ----------

func (a *API) AdminListUsers(w http.ResponseWriter, r *http.Request) {
	list, err := a.Users.List(r.Context(), strings.TrimSpace(r.URL.Query().Get("search")))
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kullanicilar listelenemedi")
		return
	}
	JSON(w, http.StatusOK, list)
}

// ---------- Kategori yonetimi ----------

func (a *API) AdminCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := DecodeBody(r, &req); err != nil || req.Name == "" || req.Slug == "" {
		ErrorJSON(w, http.StatusBadRequest, "ad ve slug zorunlu")
		return
	}
	c, err := a.Tours.CreateCategory(r.Context(), req.Name, req.Slug, req.Description)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kategori olusturulamadi")
		return
	}
	JSON(w, http.StatusCreated, c)
}

func (a *API) AdminUpdateCategory(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := DecodeBody(r, &req); err != nil || req.Name == "" {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz kategori verisi")
		return
	}
	if err := a.Tours.UpdateCategory(r.Context(), r.PathValue("id"), req.Name, req.Slug, req.Description); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kategori guncellenemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------- Kupon yonetimi ----------

func (a *API) AdminListCoupons(w http.ResponseWriter, r *http.Request) {
	list, err := a.Coupons.List(r.Context())
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kuponlar listelenemedi")
		return
	}
	JSON(w, http.StatusOK, list)
}

func (a *API) AdminCreateCoupon(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code          string  `json:"code"`
		DiscountType  string  `json:"discount_type"` // percent | fixed
		DiscountValue float64 `json:"discount_value"`
		ValidFrom     string  `json:"valid_from"`
		ValidUntil    string  `json:"valid_until"`
	}
	if err := DecodeBody(r, &req); err != nil || req.Code == "" || req.DiscountValue <= 0 {
		ErrorJSON(w, http.StatusBadRequest, "kod ve pozitif indirim degeri zorunlu")
		return
	}
	if req.DiscountType != "percent" && req.DiscountType != "fixed" {
		req.DiscountType = "percent"
	}
	if req.DiscountType == "percent" && req.DiscountValue > 100 {
		ErrorJSON(w, http.StatusBadRequest, "yuzde indirim 100'u asamaz")
		return
	}
	c, err := a.Coupons.Create(r.Context(), strings.ToUpper(strings.TrimSpace(req.Code)),
		req.DiscountType, req.DiscountValue, req.ValidFrom, req.ValidUntil)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kupon olusturulamadi (kod benzersiz olmali)")
		return
	}
	JSON(w, http.StatusCreated, c)
}

func (a *API) AdminSetCouponActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Active bool `json:"active"`
	}
	if err := DecodeBody(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz istek")
		return
	}
	if err := a.Coupons.SetActive(r.Context(), r.PathValue("id"), req.Active); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kupon guncellenemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) AdminDeleteCoupon(w http.ResponseWriter, r *http.Request) {
	if err := a.Coupons.Delete(r.Context(), r.PathValue("id")); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kupon silinemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------- Site ayarlari ----------

// PublicSettings, giris gerektirmeyen yayin ayarlarini dondurur (anasayfa icin).
func (a *API) PublicSettings(w http.ResponseWriter, r *http.Request) {
	n, _ := a.Tours.GetSetting(r.Context(), "homepage_featured_count")
	if n == "" {
		n = "6"
	}
	JSON(w, http.StatusOK, map[string]string{"homepage_featured_count": n})
}

// AdminGetSettings, tum ayar satirlarini dondurur.
func (a *API) AdminGetSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := a.Tours.Query(r.Context(), `SELECT key, value FROM settings ORDER BY key`)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "ayarlar okunamadi")
		return
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if rows.Scan(&k, &v) == nil {
			out[k] = v
		}
	}
	JSON(w, http.StatusOK, out)
}

// AdminSetSetting, tek bir ayari yazar (key zorunlu).
func (a *API) AdminSetSetting(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := DecodeBody(r, &req); err != nil || req.Key == "" {
		ErrorJSON(w, http.StatusBadRequest, "key ve value zorunlu")
		return
	}
	if err := a.Tours.SetSetting(r.Context(), req.Key, req.Value); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "ayar kaydedilemedi")
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
