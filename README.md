# Gezgin Tur — Seyahat Acentasi Platformu

Musterilerin tur arayip inceledigi, kupon indirimiyle iyzico (3D Secure + taksit) odeme
yaptigi; admin panelinden tur/rezervasyon/odeme/kupon yonetimi yapan tam kapsamli platform.

## Mimari
- Backend: Go (chi router, katmanli: handler -> service -> repository), PostgreSQL (pgx)
- Frontend: Next.js 14 App Router + Tailwind (SSR/SSG, SEO metadata)
- Admin: Next.js /admin route grubu, rol bazli erisim (JWT + RequireAdmin)
- Odeme: iyzico Checkout Form (kart verisi sunucuda saklanmaz/loglanmaz)
- E-posta: SMTP (transactional) — Gorseller S3/MinIO imzali yuklemeyle object storage'a
- 2FA: TOTP (RFC 6238, stdlib — Google Authenticator uyumlu), login'de zorunlu

## Ozellikler
1. **Kupon sistemi:** yuzde/sabit indirim, gecerlilik tarihi, admin CRUD; rezervasyonda otomatik uygulanir
2. **S3 imzali yukleme:** admin tur formundan dogrudan S3/MinIO'ya gorsel yukleme (sunucu trafigi yok)
3. **2FA (TOTP):** Hesabim'dan kurulum (QR/otpauth URL), login'de 6 haneli kod zorunlu
4. **Admin tur formu:** dinamik dahil/haric/getirilecekler/gun-programi/gorsel/kalkis yonetimi

## Odeme akisi
1. Rezervasyon olusturulur -> durum pending, odeme kaydi initiated (kupon indirimi bu adimda uygulanir)
2. POST /api/v1/payments/checkout/init -> iyzico form HTML'i frontend'e gomulur (2/3/6/9 taksit, 3DS)
3. iyzico POST /api/v1/payments/callback + webhook ile sonuc doner -> paid/failed + onay e-postasi
4. Admin panelinden iyzico Refund API ile iade -> refunded + kontenjan iadesi

## Guvenlik
- JWT access (15 dk) + refresh token (7 gun, hash'lenmis saklanir, rotation'li)
- 2FA (TOTP) opsiyonel; aciksa login'de tek kullanimlik kod zorunlu (±1 zaman penceresi toleransli)
- Kimlik/pasaport no AES-256-GCM sifreli (KVKK); yalnizca son 4 hane maskeli gosterilir
- Imzali yuklemede yalnizca JPEG/PNG/WebP/GIF MIME turlerine izin verilir; anahtar rastgele uretilir
- Tum API'de IP bazli rate limiting; admin rotalarinda rol kontrolu
- Tum gizli degerler ortam degiskeninde (.env repoya girmez)

## Kurulum
docker compose up -d                    # PostgreSQL + MailHog
cd backend && cp ../.env.example .env   # iyzico + S3 anahtarlarini gir
go mod tidy && go run ./cmd/server      # :8080
go run ./cmd/seed                       # admin@example.com / Admin123!
cd frontend && npm install && npm run dev  # :3000

## Test
- iyzico sandbox: kart 5526080000000006, SK 12/30, CVC 123
- E-posta: MailHog arayuzu http://localhost:8025
- MinIO (opsiyonel gorsel yukleme testi): docker run -p 9000:9000 -p 9001:9001 minio/minio server /data
- 2FA testi: Google Authenticator'a otpauth URL'ini taratip Hesabim'dan kodu dogrula
