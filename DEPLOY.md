# Cloudflare'e Otomatik Dagitim (Push -> Yayin)

Bu proje, Git'a atildiginda **frontend'i Cloudflare Workers'a otomatik dagitan**
bir GitHub Actions akisi iceriyor (`.github/workflows/deploy-cloudflare.yml`).

> Onemli: Frontend Cloudflare'de calisir, ancak **Go backend'i Cloudflare'de barinamaz**.
> API'yi herkese acik bir adrese kurmaniz gerekir (VPS + systemd, Railway/Render ucretsiz gibi).
> Canlida tur detaylari SSR ile cekildigi icin API URL'i public olmalidir.

## Adim adim kurulum

### 1) Backend'i internete acin
Ornek VPS (Ubuntu):
```bash
cd seyahat-acentasi/backend
# .env'e PUBLIC_BASE_URL=https://sizinsiteniz.com, CORS, DB ve iyzico anahtarlarini girin
gcloud=false # (istemeyin)
sudo tee /etc/systemd/system/seyahat-api.service <<EOF
[Unit]
Description=Gezgin Tur API
After=network.target
[Service]
WorkingDirectory=/opt/seyahat/backend
ExecStart=/usr/local/go/bin/go run ./cmd/server
Restart=always
User=www-data
[Install]
WantedBy=multi-user.target
EOF
systemctl enable --now seyahat-api
# Nginx/traefik ile https://api.sizinsiteniz.com -> :8080 yonlendirin
```

### 2) Cloudflare hesap bilgilerini GitHub'a ekleyin
Repo: **Settings → Secrets and variables → Actions**
- Secret `CLOUDFLARE_API_TOKEN` (Workers Edit izni; dash.cloudflare.com → My Profile → API Tokens)
- Secret `CLOUDFLARE_ACCOUNT_ID`
- Variable `NEXT_PUBLIC_API_URL` = `https://api.sizinsiteniz.com/api/v1`
- Variable `API_INTERNAL_URL` = `https://api.sizinsiteniz.com` (frontend rewrite hedefi)
- (Opsiyonel) `NEXT_PUBLIC_GOOGLE_CLIENT_ID`, `NEXT_PUBLIC_APPLE_SERVICE_ID`

### 3) Wrangler ayari (frontend/wrangler.jsonc)
Workers servis adi bu dosyadan okunur; ilk deploy'da otomatik olusur.

### 4) Push → Otomatik yayin
```bash
git add -A && git commit -m "..." && git push
```
Actions sekmesinden "Deploy to Cloudflare" adimini takip edin; bitince
`https://seyahat-web.<hesap>.workers.dev` (veya bagladiginiz alan adi) guncellenir.

## Alternatif: Cloudflare Pages (Git Entegrasyonu, kod gerektirmez)
1. dash.cloudflare.com → **Workers & Pages → Create → Pages → Connect to Git** → `mytourguide/GezginTur` secin
2. Framework: **Next.js**, Build command: `npm run build`, root: `frontend`
3. Env: `NEXT_PUBLIC_API_URL` + `API_INTERNAL_URL` tanimlayin
4. Her `git push`'da Cloudflare kendisi derleyip yayina alir.

Ikinci secenek daha basit; Pages depoya kendisi kuruldugunda workflow'a gerek kalmaz
(yine de git araciligiyla otomatik olur — Pages push'u dinler).
