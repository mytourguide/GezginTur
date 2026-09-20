-- 0005_featured_order.sql — anasayfa duzeni: tur yayin sirasi + ayarlar tablosu
ALTER TABLE tours ADD COLUMN IF NOT EXISTS publish_order INT NOT NULL DEFAULT 0;

-- Basit anahtar-deger ayar tablosu (or. anasayfada kac tur gosterilecegi)
CREATE TABLE IF NOT EXISTS settings (
    key   VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO settings (key, value) VALUES ('homepage_featured_count', '6')
ON CONFLICT (key) DO NOTHING;
