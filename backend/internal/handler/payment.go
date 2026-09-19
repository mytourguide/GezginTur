package handler

import (
	"fmt"
	"net/http"

	"seyahat/backend/internal/service"
)

// InitCheckout, rezervasyon icin iyzico Checkout Form baslatir.
// Donen checkoutFormContent, frontend tarafindan dogrudan sayfaya gomulur;
// kart bilgileri iyzico'nun tokenization altyapisinda islenir (bizde saklanmaz/loglanmaz).
func (a *API) InitCheckout(w http.ResponseWriter, r *http.Request) {
	claims := service.ClaimsFrom(r.Context())
	var req struct {
		BookingID string `json:"booking_id"`
	}
	if err := DecodeBody(r, &req); err != nil || req.BookingID == "" {
		ErrorJSON(w, http.StatusBadRequest, "booking_id zorunlu")
		return
	}
	b, err := a.Bookings.GetByID(r.Context(), req.BookingID)
	if err != nil || b == nil {
		ErrorJSON(w, http.StatusNotFound, "rezervasyon bulunamadi")
		return
	}
	if b.UserID != claims.UserID {
		ErrorJSON(w, http.StatusForbidden, "bu rezervasyon size ait degil")
		return
	}
	if b.Status != "pending" {
		ErrorJSON(w, http.StatusConflict, "bu rezervasyon icin odeme baslatilamaz (durum: "+b.Status+")")
		return
	}
	u, err := a.Users.FindByID(r.Context(), claims.UserID)
	if err != nil || u == nil {
		ErrorJSON(w, http.StatusNotFound, "kullanici bulunamadi")
		return
	}

	result, err := a.Payment.InitCheckout(r.Context(), b, u, b.TourTitle)
	if err != nil {
		ErrorJSON(w, http.StatusBadGateway, "odeme formu baslatilamadi: "+err.Error())
		return
	}
	JSON(w, http.StatusOK, result)
}

// PaymentCallback, iyzico 3D Secure/checkout sonrasi token ile geri doner.
// Token ile odeme sonucu sorgulanir; rezervasyon durumu buna gore guncellenir
// ve kullanici sonuc sayfasina yonlendirilir.
func (a *API) PaymentCallback(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz callback")
		return
	}
	token := r.FormValue("token")
	if token == "" {
		http.Redirect(w, r, a.Cfg.PublicBaseURL+"/odeme/sonuc?status=failed&reason=token-yok", http.StatusSeeOther)
		return
	}

	res, err := a.Payment.RetrieveCheckout(r.Context(), token)
	if err != nil {
		http.Redirect(w, r, a.Cfg.PublicBaseURL+"/odeme/sonuc?status=failed&reason=sorgu-hatasi", http.StatusSeeOther)
		return
	}

	// Rezervasyonu bulmak icin once Retrieve yanitindaki basketId (booking id) kullanilir;
	// bossa callback form alanlarina dusulur.
	bookingID := res.BasketID
	if bookingID == "" {
		bookingID = r.FormValue("conversationId")
	}
	if bookingID == "" {
		bookingID = r.FormValue("basketId")
	}
	a.finalizePayment(r, bookingID, res)

	if res.Success {
		http.Redirect(w, r, a.Cfg.PublicBaseURL+"/odeme/sonuc?status=success", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("%s/odeme/sonuc?status=failed&reason=%s", a.Cfg.PublicBaseURL, res.ErrorMsg), http.StatusSeeOther)
}

// finalizePayment, odeme sonucuna gore payment + booking kayitlarini gunceller
// ve musteriye e-posta gonderir.
func (a *API) finalizePayment(r *http.Request, bookingID string, res *service.PaymentResult) {
	ctx := r.Context()
	p, err := a.Payments.GetByBookingID(ctx, bookingID)
	if err != nil || p == nil {
		return
	}
	newStatus := "failed"
	bookingStatus := "failed"
	if res.Success {
		newStatus = "success"
		bookingStatus = "paid"
	}
	_ = a.Payments.UpdateStatus(ctx, p.ID, newStatus, res.PaymentID, res.Installment)
	_ = a.Bookings.UpdateStatus(ctx, bookingID, bookingStatus)

	b, _ := a.Bookings.GetByID(ctx, bookingID)
	if b != nil {
		u, _ := a.Users.FindByID(ctx, b.UserID)
		if u != nil {
			if res.Success {
				_ = a.Email.SendBookingConfirmation(u.Email, u.FullName, b.TourTitle, b.ID, b.TotalPrice)
			} else {
				_ = a.Email.SendPaymentFailed(u.Email, b.TourTitle)
			}
		}
	}
}

// PaymentWebhook, iyzico'nun sunucudan sunucuya bildirimlerini alir.
// Webhook URL'sine merchant panelinde secret query parametresi eklenerek
// kaynagin dogrulanmasi saglanir: /api/v1/payments/webhook?key=SIRLI_DEGER
func (a *API) PaymentWebhook(w http.ResponseWriter, r *http.Request) {
	if a.Cfg.WebhookSecret != "" && r.URL.Query().Get("key") != a.Cfg.WebhookSecret {
		ErrorJSON(w, http.StatusUnauthorized, "gecersiz webhook anahtari")
		return
	}
	var payload struct {
		IyziEventType         string `json:"iyziEventType"`         // API_AUTH | THREE_DS_AUTH | BKM_AUTH
		IyziReferenceCode     string `json:"iyziReferenceCode"`
		PaymentID             string `json:"paymentId"`
		PaymentConversationID string `json:"paymentConversationId"` // bizim conversationId = booking id
		Status                string `json:"status"`                  // SUCCESS | FAILURE
	}
	if err := DecodeBody(r, &payload); err != nil || payload.PaymentID == "" {
		ErrorJSON(w, http.StatusBadRequest, "gecersiz webhook govdesi")
		return
	}

	newStatus := "failed"
	if payload.Status == "SUCCESS" {
		newStatus = "success"
	}
	_ = a.Payments.UpdateStatusByIyzicoID(r.Context(), payload.PaymentID, newStatus)
	_ = a.Bookings.UpdateStatus(r.Context(), payload.PaymentConversationID, map[string]string{
		"success": "paid", "failed": "failed",
	}[newStatus])

	w.WriteHeader(http.StatusOK) // iyzico 200 bekler; aksi halde tekrar dener
}
