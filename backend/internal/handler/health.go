package handler

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

)

// ---------- Sistem Sagligi: tek uctan tum modullerin kontrolu ----------

type healthCheck struct {
	Group   string `json:"group"`
	Name    string `json:"name"`
	Status  string `json:"status"` // ok | warn | fail
	Detail  string `json:"detail"`
	RunMs   int64  `json:"ms"`
}

// check, zamanlamali tek kontrol calistirir.
func check(group, name string, fn func() (string, string)) healthCheck {
	start := time.Now()
	status, detail := fn()
	return healthCheck{Group: group, Name: name, Status: status, Detail: detail, RunMs: time.Since(start).Milliseconds()}
}

// AdminHealth, tum sistem modullerini kontrol edip sonuclari dondurur.
// ?group= parametresiyle tek grup da calistirilabilir.
func (a *API) AdminHealth(w http.ResponseWriter, r *http.Request) {
	group := r.URL.Query().Get("group")
	ctx := r.Context()
	wanted := func(g string) bool { return group == "" || group == g }

	var out []healthCheck

	// --- Altyapi ---
	if wanted("altyapi") {
		out = append(out,
			check("altyapi", "PostgreSQL baglantisi", func() (string, string) {
				c, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()
				var one int
				if err := a.Users.QueryRow(c, "SELECT 1").Scan(&one); err != nil {
					return "fail", err.Error()
				}
				return "ok", "baglanti aktif"
			}),
			check("altyapi", "JWT gizli anahtari", func() (string, string) {
				if strings.Contains(a.Cfg.JWTSecret, "dev-") || len(a.Cfg.JWTSecret) < 16 {
					return "warn", "JWT_SECRET gelistirme degeri kullaniliyor"
				}
				return "ok", "ozel anahtar ayarli"
			}),
			check("altyapi", "SMTP (e-posta)", func() (string, string) {
				conn, err := net.DialTimeout("tcp", net.JoinHostPort(a.Cfg.SMTPHost, a.Cfg.SMTPPort), 3*time.Second)
				if err != nil {
					return "fail", a.Cfg.SMTPHost + ":" + a.Cfg.SMTPPort + " ulasilamaz (" + err.Error() + ")"
				}
				conn.Close()
				return "ok", a.Cfg.SMTPHost + ":" + a.Cfg.SMTPPort + " erisilebilir"
			}),
		)
	}

	// --- Kimlik / guvenlik ---
	if wanted("guvenlik") {
		out = append(out,
			check("guvenlik", "Admin hesaplari", func() (string, string) {
				var n int
				a.Users.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role='admin'`).Scan(&n)
				if n == 0 {
					return "fail", "admin kullanici yok"
				}
				return "ok", itoa(n) + " admin"
			}),
			check("guvenlik", "Admin 2FA durumu", func() (string, string) {
				var n int
				a.Users.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role='admin' AND totp_enabled=true`).Scan(&n)
				if n == 0 {
					return "warn", "hicbir adminde 2FA aktif degil"
				}
				return "ok", itoa(n) + " admin 2FA kullaniyor"
			}),
			check("guvenlik", "Kart sifreleme anahtari (KVKK)", func() (string, string) {
				if a.Cfg.DataEncryptionKey == "" {
					return "warn", "DATA_ENCRYPTION_KEY bos (kimlik/pasaport sifreleme kapali)"
				}
				return "ok", "ayarli"
			}),
		)
	}

	// --- Icerik (tur/veri butunlugu) ---
	if wanted("icerik") {
		out = append(out,
			check("icerik", "Aktif turlar", func() (string, string) {
				var n int
				a.Tours.QueryRow(ctx, `SELECT COUNT(*) FROM tours WHERE active=true`).Scan(&n)
				if n == 0 {
					return "fail", "aktif tur yok"
				}
				return "ok", itoa(n) + " aktif tur"
			}),
			check("icerik", "Kapak gorseli olmayan turlar", func() (string, string) {
				var n int
				a.Tours.QueryRow(ctx, `SELECT COUNT(*) FROM tours WHERE active=true AND cover_image=''`).Scan(&n)
				if n > 0 {
					return "warn", itoa(n) + " turun kapak gorseli eksik"
				}
				return "ok", "tum aktif turlarda gorsel var"
			}),
			check("icerik", "EN ceviri eksikligi", func() (string, string) {
				var n int
				a.Tours.QueryRow(ctx, `SELECT COUNT(*) FROM tours WHERE active=true AND (title_en IS NULL OR title_en='' OR description_en IS NULL OR description_en='')`).Scan(&n)
				if n > 0 {
					return "warn", itoa(n) + " turde EN ceviri eksik"
				}
				return "ok", "tum turlar iki dilli"
			}),
			check("icerik", "Gelecek kalkis tarihleri", func() (string, string) {
				var noDep, tot int
				a.Tours.QueryRow(ctx, `SELECT COUNT(*), COUNT(*) FILTER (WHERE EXISTS (SELECT 1 FROM tour_departures d WHERE d.tour_id=tours.id AND d.start_date >= CURRENT_DATE)) FROM tours WHERE active=true`).Scan(&tot, &noDep)
				if tot > 0 && noDep == 0 {
					return "warn", "hicbir turun gelecek kalkis tarihi yok"
				}
				return "ok", itoa(noDep) + "/" + itoa(tot) + " turde gelecek tarih var"
			}),
		)
	}

	// --- Odeme ---
	if wanted("odeme") {
		out = append(out,
			check("odeme", "iyzico bilgileri", func() (string, string) {
				if a.Cfg.IyzicoAPIKey == "" || a.Cfg.IyzicoSecretKey == "" {
					return "fail", "IYZICO API anahtarlari tanimli degil"
				}
				mode := "PRODUCTION"
				if strings.Contains(a.Cfg.IyzicoBaseURL, "sandbox") {
					mode = "SANDBOX"
				}
				return "ok", "mod: " + mode
			}),
			check("odeme", "Bekleyen eski rezervasyonlar", func() (string, string) {
				var n int
				a.Bookings.QueryRow(ctx, `SELECT COUNT(*) FROM bookings WHERE status='pending' AND created_at < now() - interval '24 hours'`).Scan(&n)
				if n > 0 {
					return "warn", itoa(n) + " rezervasyon 24 saattir pending"
				}
				return "ok", "aski bekleyen odeme yok"
			}),
			check("odeme", "Basarisiz odemeler", func() (string, string) {
				var n int
				a.Payments.QueryRow(ctx, `SELECT COUNT(*) FROM payments WHERE status='failed'`).Scan(&n)
				if n > 0 {
					return "warn", itoa(n) + " basarisiz odeme kaydi var"
				}
				return "ok", "basarisiz odeme yok"
			}),
		)
	}

	// --- Saklama ---
	if wanted("saklama") {
		out = append(out,
			check("saklama", "S3/object storage", func() (string, string) {
				if a.Cfg.S3Endpoint == "" || a.Cfg.S3Bucket == "" {
					return "warn", "S3 ayarli degil (gorsel yukleme imzalari kapali)"
				}
				return "ok", "bucket: " + a.Cfg.S3Bucket
			}),
		)
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"checks":       out,
		"summary": map[string]int{
			"ok":   sumStatus(out, "ok"),
			"warn": sumStatus(out, "warn"),
			"fail": sumStatus(out, "fail"),
		},
	})
}

func sumStatus(list []healthCheck, s string) int {
	n := 0
	for _, c := range list {
		if c.Status == s {
			n++
		}
	}
	return n
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
