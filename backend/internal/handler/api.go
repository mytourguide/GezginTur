package handler

import (
	"seyahat/backend/internal/config"
	"seyahat/backend/internal/repository"
	"seyahat/backend/internal/service"
)

// API, handler katmaninin merkezidir; tum bagimliliklari tutar.
// main.go bu yapiyi doldurur, router.go metodlarini route'lara baglar.
type API struct {
	Cfg      *config.Config
	Users    *repository.UserRepository
	Tours    *repository.TourRepository
	Bookings *repository.BookingRepository
	Payments *repository.PaymentRepository
	Coupons  *repository.CouponRepository
	Tokens   *service.TokenService
	Email    *service.EmailService
	Crypto   *service.CryptoService
	Payment  *service.PaymentService
	Uploads  *service.UploadService
}
