ALTER TABLE photos ADD COLUMN featured BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX idx_photos_featured ON photos(featured) WHERE featured = true;
