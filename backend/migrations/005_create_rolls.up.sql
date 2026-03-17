-- 1. Create rolls table (without cover_photo_id FK initially)
CREATE TABLE IF NOT EXISTS rolls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    camera TEXT,
    film_stock TEXT,
    lens TEXT,
    location TEXT,
    shot_at TIMESTAMPTZ,
    published BOOLEAN NOT NULL DEFAULT false,
    cover_photo_id UUID,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_rolls_slug ON rolls(slug);
CREATE INDEX idx_rolls_published ON rolls(published, sort_order, created_at DESC);

-- Auto-update updated_at trigger
CREATE TRIGGER set_rolls_updated_at
    BEFORE UPDATE ON rolls
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- 2. Add roll_id as nullable, location, and hidden to photos
ALTER TABLE photos ADD COLUMN roll_id UUID;
ALTER TABLE photos ADD COLUMN location TEXT;
ALTER TABLE photos ADD COLUMN hidden BOOLEAN NOT NULL DEFAULT false;

-- 3. Create default "Imported" roll for existing photos
INSERT INTO rolls (id, title, slug, published, sort_order)
SELECT gen_random_uuid(), 'Imported', 'imported', true, 0
WHERE EXISTS (SELECT 1 FROM photos LIMIT 1);

-- 4. Set roll_id on all existing photos
UPDATE photos SET roll_id = (SELECT id FROM rolls WHERE slug = 'imported')
WHERE roll_id IS NULL;

-- 5. Set hidden = NOT published on existing photos
UPDATE photos SET hidden = NOT published;

-- 6. Make roll_id NOT NULL and add FK constraint
ALTER TABLE photos ALTER COLUMN roll_id SET NOT NULL;
ALTER TABLE photos ADD CONSTRAINT fk_photos_roll
    FOREIGN KEY (roll_id) REFERENCES rolls(id) ON DELETE CASCADE;

-- 7. Drop published column
ALTER TABLE photos DROP COLUMN published;

-- 8. Make title nullable (already nullable, but ensure)
ALTER TABLE photos ALTER COLUMN title DROP NOT NULL;

-- 9. Add cover_photo_id FK to rolls
ALTER TABLE rolls ADD CONSTRAINT fk_rolls_cover_photo
    FOREIGN KEY (cover_photo_id) REFERENCES photos(id) ON DELETE SET NULL;

-- 10. Update photo indexes (replace published-based index)
DROP INDEX IF EXISTS idx_photos_published;
CREATE INDEX idx_photos_roll_id ON photos(roll_id, sort_order, created_at DESC);
CREATE INDEX idx_photos_hidden ON photos(hidden, created_at DESC);
