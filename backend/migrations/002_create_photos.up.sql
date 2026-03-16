CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE photos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT,
    description TEXT,
    slug TEXT NOT NULL UNIQUE,
    film_stock TEXT,
    camera TEXT,
    lens TEXT,
    taken_at TIMESTAMPTZ,
    published BOOLEAN NOT NULL DEFAULT false,
    original_key TEXT NOT NULL,
    variants JSONB NOT NULL DEFAULT '{}',
    width INT NOT NULL,
    height INT NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    blur_hash TEXT,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER photos_updated_at
    BEFORE UPDATE ON photos
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_photos_published ON photos (published, sort_order, created_at DESC);
CREATE INDEX idx_photos_slug ON photos (slug);
