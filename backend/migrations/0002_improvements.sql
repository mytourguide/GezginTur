-- 0002_improvements.sql — Kupon indirimi, S3 ve 2FA (TOTP) destegi

-- Rezervasyonlara kupon/indirim kolonlari
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS coupon_code      VARCHAR(50);
ALTER TABLE bookings ADD COLUMN IF NOT EXISTS discount_amount  NUMERIC(12,2) NOT NULL DEFAULT 0;

-- Kullanicilara TOTP (2FA) kolonlari
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret  VARCHAR(64);
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled BOOLEAN NOT NULL DEFAULT false;
