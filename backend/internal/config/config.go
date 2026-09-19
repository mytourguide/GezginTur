package config

import "os"

// Config, uygulamanin tum ortam degiskenlerini tutar.
// Gizli degerler asla koda gomulmez; yalnizca .env / ortam degiskeninden okunur.
type Config struct {
	Addr              string
	DatabaseURL       string
	JWTSecret         string
	DataEncryptionKey string // KVKK: pasaport/kimlik no sifreleme anahtari (32 byte, base64)
	PublicBaseURL     string // Odeme callback'lerinde kullanilan dis adres
	WebhookSecret     string // iyzico webhook dogrulama anahtari (URL ?key=...)

	// iyzico
	IyzicoAPIKey    string
	IyzicoSecretKey string
	IyzicoBaseURL   string // sandbox: https://sandbox-api.iyzipay.com

	// S3 uyumlu object storage (imzali yukleme icin)
	S3Endpoint  string
	S3Region    string
	S3Bucket    string
	S3AccessKey string
	S3SecretKey string
	S3PublicURL string // dosyalarin erisilebilir oldugu taban URL

	// SMTP (transactional e-posta)
	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// Sosyal giris (OAuth ID token dogrulamasinda audience kontrolu)
	GoogleClientID string
	AppleServiceID string
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load, ortam degiskenlerini okuyup Config dondurur.
func Load() *Config {
	return &Config{
		Addr:              env("ADDR", ":8080"),
		DatabaseURL:       env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/seyahat?sslmode=disable"),
		JWTSecret:         env("JWT_SECRET", "dev-secret-degistir"),
		DataEncryptionKey: env("DATA_ENCRYPTION_KEY", ""),
		PublicBaseURL:     env("PUBLIC_BASE_URL", "http://localhost:3000"),
		WebhookSecret:     env("WEBHOOK_SECRET", ""),

		IyzicoAPIKey:    env("IYZICO_API_KEY", ""),
		IyzicoSecretKey: env("IYZICO_SECRET_KEY", ""),
		IyzicoBaseURL:   env("IYZICO_BASE_URL", "https://sandbox-api.iyzipay.com"),

		S3Endpoint:  env("S3_ENDPOINT", ""), // MinIO icin ornek: http://localhost:9000
		S3Region:    env("S3_REGION", "eu-central-1"),
		S3Bucket:    env("S3_BUCKET", ""),
		S3AccessKey: env("S3_ACCESS_KEY", ""),
		S3SecretKey: env("S3_SECRET_KEY", ""),
		S3PublicURL: env("S3_PUBLIC_URL", ""),

		SMTPHost: env("SMTP_HOST", "localhost"),
		SMTPPort: env("SMTP_PORT", "1025"),
		SMTPUser: env("SMTP_USER", ""),
		SMTPPass: env("SMTP_PASS", ""),
		SMTPFrom: env("SMTP_FROM", "rezervasyon@example.com"),

		GoogleClientID: env("GOOGLE_CLIENT_ID", ""),
		AppleServiceID: env("APPLE_SERVICE_ID", ""),
	}
}
