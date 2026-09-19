-- 0001_init.sql — Seyahat acentasi veritabani semasi (PostgreSQL)

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Kullanicilar (musteri + admin)
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name     VARCHAR(150) NOT NULL,
    email         VARCHAR(255) NOT NULL UNIQUE,
    phone         VARCHAR(30),
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(20)  NOT NULL DEFAULT 'customer' CHECK (role IN ('customer', 'admin')),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Refresh token'lar (token'in kendisi degil, SHA-256 hash'i saklanir)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);

-- Tur kategorileri
CREATE TABLE IF NOT EXISTS tour_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(120) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Turlar
CREATE TABLE IF NOT EXISTS tours (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           VARCHAR(255) NOT NULL,
    slug            VARCHAR(255) NOT NULL UNIQUE,
    description     TEXT NOT NULL DEFAULT '',
    category_id     UUID NOT NULL REFERENCES tour_categories(id),
    cover_image     VARCHAR(500) NOT NULL DEFAULT '',
    duration_days   INT  NOT NULL DEFAULT 1,
    duration_nights INT  NOT NULL DEFAULT 0,
    location        VARCHAR(255) NOT NULL DEFAULT '',
    base_price      NUMERIC(12,2) NOT NULL DEFAULT 0,
    currency        CHAR(3) NOT NULL DEFAULT 'TRY',
    active          BOOLEAN NOT NULL DEFAULT true,
    featured        BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_tours_category ON tours(category_id);
CREATE INDEX IF NOT EXISTS idx_tours_active ON tours(active);

-- Tur gorselleri
CREATE TABLE IF NOT EXISTS tour_images (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id    UUID NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    image_url  VARCHAR(500) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_tour_images_tour ON tour_images(tour_id);

-- Kalkis tarihleri (kontenjan + tarih bazli fiyat)
CREATE TABLE IF NOT EXISTS tour_departures (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id    UUID NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date   DATE NOT NULL,
    capacity   INT  NOT NULL CHECK (capacity > 0),
    filled     INT  NOT NULL DEFAULT 0 CHECK (filled >= 0),
    price      NUMERIC(12,2) NOT NULL CHECK (price >= 0)
);
CREATE INDEX IF NOT EXISTS idx_departures_tour ON tour_departures(tour_id);
CREATE INDEX IF NOT EXISTS idx_departures_start ON tour_departures(start_date);

-- Dahil olanlar / olmayanlar / yaninizda getirmeniz gerekenler
CREATE TABLE IF NOT EXISTS tour_included_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id     UUID NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    description VARCHAR(500) NOT NULL
);
CREATE TABLE IF NOT EXISTS tour_excluded_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id     UUID NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    description VARCHAR(500) NOT NULL
);
CREATE TABLE IF NOT EXISTS tour_bring_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id     UUID NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    description VARCHAR(500) NOT NULL
);

-- Gun gun program
CREATE TABLE IF NOT EXISTS tour_itinerary (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id     UUID NOT NULL REFERENCES tours(id) ON DELETE CASCADE,
    day_no      INT NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_itinerary_tour ON tour_itinerary(tour_id, day_no);

-- Rezervasyonlar
CREATE TABLE IF NOT EXISTS bookings (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tour_id      UUID NOT NULL REFERENCES tours(id),
    departure_id UUID NOT NULL REFERENCES tour_departures(id),
    user_id      UUID NOT NULL REFERENCES users(id),
    adult_count  INT  NOT NULL CHECK (adult_count > 0),
    child_count  INT  NOT NULL DEFAULT 0 CHECK (child_count >= 0),
    total_price  NUMERIC(12,2) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'paid', 'confirmed', 'cancelled', 'failed')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_status ON bookings(status);
CREATE INDEX IF NOT EXISTS idx_bookings_created ON bookings(created_at);

-- Yolcular — kimlik/pasaport numarasi KVKK geregi sifreli saklanir (AES-256-GCM),
-- duz metin asla bu tabloya yazilmaz. Yalnizca son 4 hane maskeleme icin tutulur.
CREATE TABLE IF NOT EXISTS booking_travelers (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id            UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    full_name             VARCHAR(150) NOT NULL,
    birth_date            DATE NOT NULL,
    id_number_encrypted   TEXT NOT NULL,        -- base64(nonce||ciphertext)
    id_number_last4       VARCHAR(4) NOT NULL,  -- maskeleme icin son 4 hane
    is_child              BOOLEAN NOT NULL DEFAULT false
);
CREATE INDEX IF NOT EXISTS idx_travelers_booking ON booking_travelers(booking_id);

-- Odemeler (kart verisi ASLA saklanmaz; yalnizca iyzico referansi tutulur)
CREATE TABLE IF NOT EXISTS payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id        UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    iyzico_payment_id VARCHAR(64),
    amount            NUMERIC(12,2) NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'initiated'
                      CHECK (status IN ('initiated', 'success', 'failed', 'refunded', 'partial_refunded')),
    installment       INT NOT NULL DEFAULT 1,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_payments_booking ON payments(booking_id);
CREATE INDEX IF NOT EXISTS idx_payments_iyzico ON payments(iyzico_payment_id);

-- Kuponlar (opsiyonel)
CREATE TABLE IF NOT EXISTS coupons (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code           VARCHAR(50) NOT NULL UNIQUE,
    discount_type  VARCHAR(10) NOT NULL CHECK (discount_type IN ('percent', 'fixed')),
    discount_value NUMERIC(12,2) NOT NULL,
    valid_from     DATE,
    valid_until    DATE,
    active         BOOLEAN NOT NULL DEFAULT true
);

-- ---------- Ornek veri (gelistirme ortami) ----------

INSERT INTO tour_categories (id, name, slug, description) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Yurt Ici', 'yurt-ici', 'Turkiye ici destinasyonlar'),
    ('22222222-2222-2222-2222-222222222222', 'Yurt Disi', 'yurt-disi', 'Uluslararasi turlar'),
    ('33333333-3333-3333-3333-333333333333', 'Kultur Turlari', 'kultur-turlari', 'Tarih ve kultur odakli'),
    ('44444444-4444-4444-4444-444444444444', 'Doga Turlari', 'doga-turlari', 'Doga ile ic ice rotalar'),
    ('55555555-5555-5555-5555-555555555555', 'Balayi', 'balayi', 'Romantik kacamaklar')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO tours (id, title, slug, description, category_id, cover_image, duration_days, duration_nights, location, base_price, featured) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Kapadokya Balon Turu', 'kapadokya-balon-turu',
     'Peribacalari vadilerinde gun dogumu balon deneyimi, yeralti sehirleri ve canak-comlek atolyeleri ile unutulmaz bir Kapadokya kacamagi.',
     '11111111-1111-1111-1111-111111111111',
     'https://images.unsplash.com/photo-1570939274717-7eda259b50ed?w=1200&q=80',
     3, 2, 'Nevsehir, Turkiye', 7500.00, true),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Italya Ruyasi: Roma & Floransa', 'italya-ruyasi-roma-floransa',
     'Kolezyum, Vatikan, Floransa katedrali ve Toskana baglari... Italya''nin en ikonik sehirlerini rehberligimizle kesfedin.',
     '22222222-2222-2222-2222-222222222222',
     'https://images.unsplash.com/photo-1552832230-c0197dd311b5?w=1200&q=80',
     7, 6, 'Roma & Floransa, Italya', 38500.00, true)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO tour_images (tour_id, image_url, sort_order) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'https://images.unsplash.com/photo-1570939274717-7eda259b50ed?w=1200&q=80', 0),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'https://images.unsplash.com/photo-1541432901042-2d8bd64b4a9b?w=1200&q=80', 1),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'https://images.unsplash.com/photo-1552832230-c0197dd311b5?w=1200&q=80', 0),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'https://images.unsplash.com/photo-1514890547357-a9ee288728e0?w=1200&q=80', 1)
ON CONFLICT DO NOTHING;

INSERT INTO tour_departures (tour_id, start_date, end_date, capacity, price) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', CURRENT_DATE + 30, CURRENT_DATE + 32, 20, 7500.00),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', CURRENT_DATE + 60, CURRENT_DATE + 62, 20, 8200.00),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', CURRENT_DATE + 45, CURRENT_DATE + 51, 25, 38500.00)
ON CONFLICT DO NOTHING;

INSERT INTO tour_included_items (tour_id, description) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Otelde 2 gece konaklama (kahvalti dahil)'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Gun dogumu balon turu (standart sepet)'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Profesyonel Turkce rehberlik'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '4-5 yildizli otellerde 6 gece konaklama'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Ic hat ucak biletleri ve havalimani transferleri'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Sehir turlari ve muze giris ucretleri')
ON CONFLICT DO NOTHING;

INSERT INTO tour_excluded_items (tour_id, description) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Kisisel harcamalar'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Balayi sepeti yukseltmesi (ek ucretli)'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Schengen vize ucreti'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Ogle ve aksam yemekleri (belirtilenler disinda)')
ON CONFLICT DO NOTHING;

INSERT INTO tour_bring_items (tour_id, description) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Nufus cuzdani veya pasaport'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Yuruyuse uygun rahat ayakkabi'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Pasaport (6 ay gecerlilik)'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Schengen vizesi / seyahat saglik sigortasi')
ON CONFLICT DO NOTHING;

INSERT INTO tour_itinerary (tour_id, day_no, title, description) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 1, 'Goreme Vadisi & Acik Hava Muzesi', 'Sabah Kayseri havalimani karsilamasi, Goreme Acik Hava Muzesi turu ve gun batimi ATV turu.'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 2, 'Balon Turu & Yeralti Sehri', 'Gun dogumu balon deneyimi, ardindan Derinkuyu yeralti sehri ve canak-comlek atolyesi.'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 3, 'Ihlara Vadisi & Donus', 'Ihlara Vadisi yuruyusu ve Belisirma koyunde ogle yemegi sonrasi donus.'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 1, 'Romaya Varis', 'Havalimani transferi ve serbest zaman, aksam Trastevere turu.'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 2, 'Antik Roma', 'Kolezyum, Roma Forumu ve Palatin Tepesi rehberli turu.'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 7, 'Donus', 'Serbest zaman ve havalimani transferi.')
ON CONFLICT DO NOTHING;
