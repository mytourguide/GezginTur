-- 0004_coupon_created_at.sql — coupons.created_at sorgulanissa tabloda kolon yoktu (ekle)
ALTER TABLE coupons ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT now();
