DROP INDEX IF EXISTS idx_photos_featured;
ALTER TABLE photos DROP COLUMN IF EXISTS featured;
