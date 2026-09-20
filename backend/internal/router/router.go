package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	chimw "github.com/go-chi/chi/v5/middleware"

	"seyahat/backend/internal/config"
	"seyahat/backend/internal/handler"
	mw "seyahat/backend/internal/middleware"
	"seyahat/backend/internal/service"
)

// hasHTTPPrefix: origin'in http(s)://<host>... seklinde olmasini kontrol eder
func hasHTTPPrefix(origin, host string) bool {
	return strings.HasPrefix(origin, "http://"+host) || strings.HasPrefix(origin, "https://"+host)
}

// New, tum route'lari tanimlar ve handler'larla eslestirir.
func New(cfg *config.Config, api *handler.API, tokens *service.TokenService) http.Handler {
	r := chi.NewRouter()

	// Genel middleware'ler
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	// CORS: Next.js gelistirme sunucusundan API'ye istek icin gerekli.
	// PUBLIC_BASE_URL'e ek olarak localhost tabanli diger origin'lere (or. :3001) de izin ver.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")
			allowed := origin == cfg.PublicBaseURL ||
				(len(origin) > 0 && (hasHTTPPrefix(origin, "localhost:") || hasHTTPPrefix(origin, "127.0.0.1:")))
			w.Header().Set("Vary", "Origin")
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			next.ServeHTTP(w, req)
		})
	})
	r.Use(chimw.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization"))
	r.Use(chimw.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS"))
	r.Method(http.MethodOptions, "/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.RateLimit(120, time.Minute)) // genel API hiz sinirı

		// Saglik kontrolu
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Kimlik dogrulama (daha siki limit: brute-force korumasi)
		r.Route("/auth", func(r chi.Router) {
			r.Use(mw.RateLimit(20, time.Minute))
			r.Post("/register", api.Register)
			r.Post("/login", api.Login)
			r.Post("/refresh", api.Refresh)
			r.Post("/logout", api.Logout)
			r.Group(func(r chi.Router) {
				r.Use(mw.Auth(tokens))
				r.Get("/me", api.Me)
				r.Put("/me", api.UpdateProfile)
				// Iki asamali dogrulama (TOTP)
				r.Post("/2fa/setup", api.Setup2FA)
				r.Post("/2fa/enable", api.Enable2FA)
				r.Post("/2fa/disable", api.Disable2FA)
				r.Get("/2fa/status", api.TOTPStatus)
			})
		})

		// Sosyal giris (Google / Apple id_token dogrulamasi)
		r.Post("/auth/oauth/google", api.OAuthGoogle)
		r.Post("/auth/oauth/apple", api.OAuthApple)

		// Herkese acik tur verileri
		r.Get("/tours", api.ListTours)
		r.Get("/tours/{slug}", api.GetTour)
		r.Get("/categories", api.ListCategories)
		r.Get("/settings/public", api.PublicSettings) // anasayfa yayin ayarlari

		// Odeme callback/webhook: iyzico sunuculari Bearer token gondermez, ayri tutulur
		r.Post("/payments/callback", api.PaymentCallback)
		r.Post("/payments/webhook", api.PaymentWebhook)

		// Giris gerektiren musteri islemleri
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(tokens))
			r.Post("/bookings", api.CreateBooking)
			r.Get("/bookings/mine", api.MyBookings)
			r.Get("/bookings/{id}", api.GetBooking)
			r.Post("/bookings/{id}/cancel", api.CancelBooking)
			r.Post("/payments/checkout/init", api.InitCheckout)
			// S3'e dogrudan imzali yukleme (tur gorselleri vb.)
			r.Post("/uploads/sign", api.SignUpload)
		})

		// Yonetim paneli (yalnizca admin)
		r.Route("/admin", func(r chi.Router) {
			r.Use(mw.Auth(tokens))
			r.Use(mw.RequireAdmin)

			r.Get("/dashboard", api.AdminDashboard)
			r.Get("/health", api.AdminHealth) // sistem sagligi kontrol paketi
			r.Get("/settings", api.AdminGetSettings) // site ayarlari (anasayfa duzeni vb.)
			r.Put("/settings", api.AdminSetSetting)

			r.Get("/tours", api.AdminListTours)
			r.Post("/tours", api.AdminCreateTour)
			r.Get("/tours/{id}", api.AdminGetTour)
			r.Put("/tours/{id}", api.AdminUpdateTour)
			r.Delete("/tours/{id}", api.AdminDeleteTour)

			r.Post("/departures", api.AdminCreateDeparture)
			r.Delete("/departures/{id}", api.AdminDeleteDeparture)

			r.Get("/bookings", api.AdminListBookings)
			r.Get("/bookings/{id}", api.AdminGetBooking)
			r.Patch("/bookings/{id}/status", api.AdminUpdateBookingStatus)
			r.Post("/bookings/{id}/refund", api.AdminRefund)

			r.Get("/payments", api.AdminListPayments)
			r.Get("/users", api.AdminListUsers)

			r.Post("/categories", api.AdminCreateCategory)
			r.Patch("/categories/{id}", api.AdminUpdateCategory)

			r.Get("/coupons", api.AdminListCoupons)
			r.Post("/coupons", api.AdminCreateCoupon)
			r.Patch("/coupons/{id}/active", api.AdminSetCouponActive)
			r.Delete("/coupons/{id}", api.AdminDeleteCoupon)
		})
	})

	return r
}
