package handler

import (
	"math"
	"net/http"
	"strings"
	"time"

	"seyahat/backend/internal/models"
	"seyahat/backend/internal/service"
)

// Cocuk fiyat katsayisi: yetiskin fiyatinin %70'i
const childPriceRatio = 0.70

// CreateBooking, yeni rezervasyon olusturur:
// 1) Kalkis kontenjani satir kilidi (FOR UPDATE) ile kontrol edilir - yaris durumu onlenir
// 2) Toplam fiyat hesaplanir
// 3) Yolcu kimlik/pasaport numaralari KVKK geregi sifrelenerek saklanir
// 4) Durum "pending" baslar; iyzico callback'i ile "paid"/"failed" olur.
func (a *API) CreateBooking(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	var req models.CreateBookingRequest
	if err := DecodeBody(r, &req); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz istek govdesi")
		return
	}
	if req.AdultCount < 1 || req.AdultCount > 20 || req.ChildCount < 0 || req.ChildCount > 20 {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz kisi sayisi")
		return
	}
	if len(req.Travelers) != req.AdultCount+req.ChildCount {
		ErrorJSON(w, http.StatusBadRequest, "yolcu bilgileri kisi sayisiyla eslesmelidir")
		return
	}

	tx, err := a.Bookings.BeginTx(r.Context())
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "islem baslatilamadi")
		return
	}
	defer tx.Rollback(r.Context())

	dep, err := a.Tours.GetDepartureByIDForUpdate(r.Context(), tx, req.DepartureID)
	if err != nil {
		ErrorJSON(w, http.StatusNotFound, "kalkis tarihi bulunamadi")
		return
	}
	partySize := req.AdultCount + req.ChildCount
	if dep.Filled+partySize > dep.Capacity {
		ErrorJSON(w, http.StatusConflict, "secilen tarihte yeterli kontenjan yok")
		return
	}

	total := float64(req.AdultCount)*dep.Price + float64(req.ChildCount)*dep.Price*childPriceRatio
	bCouponCode := ""

	// Opsiyonel kupon indirimi: yuzde veya sabit tutar; toplam tutari asamaz.
	discount := 0.0
	if code := strings.TrimSpace(req.CouponCode); code != "" {
		coupon, err := a.Coupons.GetByCode(r.Context(), code)
		if err != nil || coupon == nil {
			ErrorJSON(w, http.StatusBadRequest, "gecersiz veya suresi dolmus kupon kodu")
			return
		}
		if coupon.DiscountType == "percent" {
			discount = total * coupon.DiscountValue / 100
		} else {
			discount = math.Min(coupon.DiscountValue, total)
		}
		bCouponCode = coupon.Code
	}
	b := &models.Booking{
		TourID: dep.TourID, DepartureID: dep.ID, UserID: claims.UserID,
		AdultCount: req.AdultCount, ChildCount: req.ChildCount,
		TotalPrice: total - discount, CouponCode: bCouponCode, DiscountAmount: discount,
	}

	travelers := make([]models.Traveler, 0, len(req.Travelers))
	for i, t := range req.Travelers {
		if t.FullName == "" || t.BirthDate == "" || t.IDNumber == "" {
			ErrorJSON(w, http.StatusBadRequest, "tum yolcu bilgileri (ad-soyad, dogum tarihi, kimlik/pasaport) zorunludur")
			return
		}
		birth, err := time.Parse("2006-01-02", t.BirthDate)
		if err != nil {
			ErrorJSON(w, http.StatusBadRequest, "dogum tarihi formati YYYY-MM-DD olmalidir")
			return
		}
		isChild := i >= req.AdultCount // listenin ilk N kaydi yetiskin
		enc, err := a.Crypto.Encrypt(t.IDNumber) // KVKK: kimlik/pasaport sifrelenir
		if err != nil {
			ErrorJSON(w, http.StatusInternalServerError, "yolcu verisi islenemedi")
			return
		}
		last4 := t.IDNumber
		if len(last4) > 4 {
			last4 = last4[len(last4)-4:]
		}
		travelers = append(travelers, models.Traveler{
			FullName: t.FullName, BirthDate: birth,
			IDNumber: enc, IDLast4: last4, IsChild: isChild,
		})
	}

	if err := a.Bookings.CreateWithTx(r.Context(), tx, b, travelers); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "rezervasyon olusturulamadi")
		return
	}
	if err := a.Bookings.FillCapacity(r.Context(), tx, dep.ID, partySize); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "kontenjan guncellenemedi")
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "islem tamamlanamadi")
		return
	}

	// Odeme kaydini "initiated" durumuyla ac; iyzico sonucu bu kaydi gunceller.
	_ = a.Payments.Create(r.Context(), &models.Payment{
		BookingID: b.ID, Amount: b.TotalPrice, Status: "initiated", Installment: 1,
	})
	JSON(w, http.StatusCreated, b)
}

// GetBooking, rezervasyon detayini dondurur (yalnizca sahibi veya admin gorebilir).
func (a *API) GetBooking(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	b, err := a.Bookings.GetByID(r.Context(), r.PathValue("id"))
	if err != nil || b == nil {
		ErrorJSON(w, http.StatusNotFound, "rezervasyon bulunamadi")
		return
	}
	if b.UserID != claims.UserID && claims.Role != "admin" {
		ErrorJSON(w, http.StatusForbidden, "bu rezervasyonu goruntuleme yetkiniz yok")
		return
	}
	JSON(w, http.StatusOK, b)
}

// MyBookings, giris yapmis musterinin rezervasyonlarini listeler.
func (a *API) MyBookings(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	list, err := a.Bookings.ListByUser(r.Context(), claims.UserID)
	if err != nil {
		ErrorJSON(w, http.StatusInternalServerError, "rezervasyonlar listelenemedi")
		return
	}
	JSON(w, http.StatusOK, list)
}

// CancelBooking, bekleyen rezervasyonu iptal eder (kontenjan iade edilir).
func (a *API) CancelBooking(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	b, err := a.Bookings.GetByID(r.Context(), r.PathValue("id"))
	if err != nil || b == nil {
		ErrorJSON(w, http.StatusNotFound, "rezervasyon bulunamadi")
		return
	}
	if b.UserID != claims.UserID {
		ErrorJSON(w, http.StatusForbidden, "bu rezervasyonu iptal etme yetkiniz yok")
		return
	}
	if err := a.Bookings.Cancel(r.Context(), b.ID); err != nil {
		ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}
