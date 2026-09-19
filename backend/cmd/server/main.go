package main

import (
	"context"
	"log"
	"net/http"

	"seyahat/backend/internal/config"
	"seyahat/backend/internal/database"
	"seyahat/backend/internal/handler"
	"seyahat/backend/internal/repository"
	"seyahat/backend/internal/router"
	"seyahat/backend/internal/service"
)

// Uygulamanin giris noktasi: bagimliliklari kurar ve HTTP sunucusunu baslatir.
func main() {
	cfg := config.Load()

	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("veritabani baglantisi kurulamadi: %v", err)
	}
	defer pool.Close()

	// Migration dosyalarini calistir (idempotent)
	if err := database.Migrate(ctx, pool, "migrations"); err != nil {
		log.Fatalf("migration calistirilamadi: %v", err)
	}

	// --- Servis katmani ---
	tokens := service.NewTokenService(cfg.JWTSecret)
	emailSvc := service.NewEmailService(cfg)
	cryptoSvc := service.NewCryptoService(cfg.DataEncryptionKey)
	paymentGW := service.NewPaymentService(cfg)
	uploadSvc := service.NewUploadService(cfg)

	// --- Repository katmani ---
	users := repository.NewUserRepository(pool)
	tours := repository.NewTourRepository(pool)
	bookings := repository.NewBookingRepository(pool)
	payments := repository.NewPaymentRepository(pool)
	coupons := repository.NewCouponRepository(pool)

	// --- Handler (API) katmani ---
	api := &handler.API{
		Cfg:      cfg,
		Users:    users,
		Tours:    tours,
		Bookings: bookings,
		Payments: payments,
		Coupons:  coupons,
		Tokens:   tokens,
		Email:    emailSvc,
		Crypto:   cryptoSvc,
		Payment:  paymentGW,
		Uploads:  uploadSvc,
	}

	r := router.New(cfg, api, tokens)
	log.Printf("Sunucu ayakta: %s (iyzico modu: %s)", cfg.Addr, cfg.IyzicoBaseURL)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
