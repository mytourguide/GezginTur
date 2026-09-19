-- 0003_i18n.sql — dil destegi (TR varsayilan, EN ceviriler)
-- Strateji: ceviri alanlari *_en kolonlarinda tutulur; API tarafi ?lang=en
-- istegi geldiginde JSON ciktisinda ceviriyi ana alana kopyalar.

ALTER TABLE tours
    ADD COLUMN IF NOT EXISTS title_en       VARCHAR(255),
    ADD COLUMN IF NOT EXISTS slug_en        VARCHAR(255),
    ADD COLUMN IF NOT EXISTS description_en TEXT,
    ADD COLUMN IF NOT EXISTS location_en    VARCHAR(255);

ALTER TABLE tour_categories
    ADD COLUMN IF NOT EXISTS name_en VARCHAR(100),
    ADD COLUMN IF NOT EXISTS slug_en VARCHAR(120);

ALTER TABLE tour_included_items
    ADD COLUMN IF NOT EXISTS description_en VARCHAR(500);

ALTER TABLE tour_excluded_items
    ADD COLUMN IF NOT EXISTS description_en VARCHAR(500);

ALTER TABLE tour_bring_items
    ADD COLUMN IF NOT EXISTS description_en VARCHAR(500);

ALTER TABLE tour_itinerary
    ADD COLUMN IF NOT EXISTS title_en       VARCHAR(255),
    ADD COLUMN IF NOT EXISTS description_en TEXT;
