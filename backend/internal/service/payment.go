package service

// Bu paket iyzico Checkout Form entegrasyonunu sarar.
// KRITIK: Kart bilgileri bizim sunucuya ASLA ugramaz; iyzico'nun tokenization
// altyapisinda islenir. Biz yalnizca odeme formunu baslatiriz ve token ile sonuc sorgulariz.

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/iliyanm/iyzipay-go/iyzipay"

	"seyahat/backend/internal/config"
	"seyahat/backend/internal/models"
)

type PaymentService struct {
	options iyzipay.Options
	baseURL string // bizim dis adres (callback uretimi icin)
}

func NewPaymentService(cfg *config.Config) *PaymentService {
	var options iyzipay.Options
	options.New(cfg.IyzicoAPIKey, cfg.IyzicoSecretKey, cfg.IyzicoBaseURL)
	return &PaymentService{options: options, baseURL: cfg.PublicBaseURL}
}

// CheckoutResult, baslatilan odeme formunun embed edilecek icerigini tasir.
type CheckoutResult struct {
	Token               string `json:"token"`
	CheckoutFormContent string `json:"checkout_form_content"` // iyzico'nun urettigi <script> HTML'i
}

// InitCheckout, iyzico Checkout Form baslatir. Donen HTML/JS musteri tarafinda
// "iyzipay-checkout-form" div'ine basilir; 3D Secure ve taksitler iyzico tarafindan yonetilir.
func (s *PaymentService) InitCheckout(ctx context.Context, b *models.Booking, user *models.User, tourTitle string) (*CheckoutResult, error) {
	// Taksit secenekleri: musteri 2/3/6/9 taksit arasindan secebilir (1 = tek cekim).
	enabledInstallments := []string{"1", "2", "3", "6", "9"}

	callbackURL := fmt.Sprintf("%s/api/v1/payments/callback", s.baseURL)

	request := iyzipay.CreateCheckoutFormInitializeRequest{
		Locale:              "tr",
		ConversationId:      b.ID,
		Price:               fmt.Sprintf("%.2f", b.TotalPrice),
		PaidPrice:           fmt.Sprintf("%.2f", b.TotalPrice),
		Currency:            "TRY",
		BasketId:            b.ID,
		PaymentGroup:        "PRODUCT",
		CallbackUrl:         callbackURL,
		EnabledInstallments: enabledInstallments,
		Buyer: iyzipay.Buyer{
			Id:                  user.ID,
			Name:                user.FullName,
			Surname:             "-",
			Email:               user.Email,
			IdentityNumber:      "11111111111", // iyzico zorunlu alan; gercek projede kullanicidan alinir
			RegistrationAddress: "Turkiye",
			City:                "Istanbul",
			Country:             "Turkey",
		},
		BillingAddress: iyzipay.Address{
			ContactName: user.FullName,
			City:        "Istanbul",
			Country:     "Turkey",
			Address:     "Fatura adresi",
		},
		BasketItems: []iyzipay.BasketItem{
			{
				Id:        b.ID,
				Name:      tourTitle,
				Category1: "Tur",
				ItemType:  "VIRTUAL",
				Price:     fmt.Sprintf("%.2f", b.TotalPrice),
			},
		},
	}

	// iyzipay-go eski API: Create hata dondurmez; ag/hata durumunda response bos olabilir.
	response := iyzipay.CheckoutFormInitialize{}.Create(request, s.options)
	if response == nil || response.Status == "" {
		return nil, errors.New("iyzico API'sine ulasilamadi")
	}
	if response.Status != "success" {
		return nil, fmt.Errorf("iyzico baslatma hatasi: %s", response.ErrorMessage)
	}
	return &CheckoutResult{
		Token:               response.Token,
		CheckoutFormContent: response.CheckoutFormContent,
	}, nil
}

// PaymentResult, checkout form sonucunun sorgulanmasiyla elde edilen bilgiler.
type PaymentResult struct {
	Success     bool
	PaymentID   string
	BasketID    string // bizim rezervasyon id (basketId)
	Price       float64
	Installment int
	CardType    string
	ErrorMsg    string
}

// RetrieveCheckout, iyzico callback sonrasi token ile odeme sonucunu sorgular.
func (s *PaymentService) RetrieveCheckout(ctx context.Context, token string) (*PaymentResult, error) {
	retrieveRequest := iyzipay.RetrieveCheckoutFormRequest{Token: token}
	resp := iyzipay.CheckoutForm{}.Retrieve(retrieveRequest, s.options)
	if resp == nil || resp.Status == "" {
		return nil, errors.New("iyzico API'sine ulasilamadi")
	}
	res := &PaymentResult{
		Success:     resp.Status == "success" && resp.PaymentStatus == "SUCCESS",
		PaymentID:   resp.PaymentId,
		BasketID:    resp.BasketId,
		CardType:    resp.CardType,
		Installment: resp.Installment,
		Price:       resp.PaidPrice,
	}
	if res.ErrorMsg == "" && !res.Success {
		res.ErrorMsg = "Odeme tamamlanamadi"
	}
	return res, nil
}

// Refund, iyzico uzerinden iade baslatir (admin panelinden tetiklenir).
func (s *PaymentService) Refund(ctx context.Context, iyzicoPaymentID string, amount float64) error {
	// Not: iyzico iade icin paymentId degil paymentTransactionId ister;
	// bu servise o bilgi payment ID olarak aktariliyor.
	request := iyzipay.CreateRefundRequest{
		Locale:               "tr",
		ConversationId:       fmt.Sprintf("refund-%d", time.Now().Unix()),
		PaymentTransactionId: iyzicoPaymentID,
		Price:                fmt.Sprintf("%.2f", amount),
		Currency:             "TRY",
	}
	resp := iyzipay.Refund{}.Create(request, s.options)
	if resp == nil || resp.Status == "" {
		return errors.New("iyzico API'sine ulasilamadi")
	}
	if resp.Status != "success" {
		return errors.New(resp.ErrorMessage)
	}
	return nil
}
