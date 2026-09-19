package service

import (
	"fmt"
	"net/smtp"

	"seyahat/backend/internal/config"
)

// EmailService, rezervasyon onayi gibi transactional e-postalari gonderir.
type EmailService struct {
	host, port, user, pass, from string
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom}
}

// SendBookingConfirmation, odeme sonrasi musteriye onay e-postasi gonderir.
func (s *EmailService) SendBookingConfirmation(to, customerName, tourTitle, bookingID string, amount float64) error {
	subject := "Rezervasyon Onayiniz"
	body := fmt.Sprintf(`<html><body>
		<h2>Rezervasyonunuz Alindi</h2>
		<p>Sayin %s,</p>
		<p><b>%s</b> turu icin rezervasyonunuz basariyla olusturuldu.</p>
		<ul>
			<li>Rezervasyon No: %s</li>
			<li>Toplam Tutar: %.2f TL</li>
		</ul>
		<p>Iyi yolculuklar dileriz!</p>
	</body></html>`, customerName, tourTitle, bookingID, amount)

	return s.send(to, subject, body)
}

// SendPaymentFailed, basarisiz odeme bildirimi gonderir.
func (s *EmailService) SendPaymentFailed(to, tourTitle string) error {
	subject := "Odeme Islemi Tamamlanamadi"
	body := fmt.Sprintf(`<html><body>
		<p>Sayin musterimiz,</p>
		<p><b>%s</b> turu icin odeme islemi tamamlanamadi. Lutfen tekrar deneyiniz.</p>
	</body></html>`, tourTitle)
	return s.send(to, subject, body)
}

func (s *EmailService) send(to, subject, html string) error {
	addr := s.host + ":" + s.port
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s",
		s.from, to, subject, html))

	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
