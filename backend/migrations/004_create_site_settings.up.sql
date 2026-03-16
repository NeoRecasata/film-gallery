CREATE TABLE site_settings (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL DEFAULT '{}'
);

INSERT INTO site_settings (key, value) VALUES
    ('site_title', '"My Gallery"'),
    ('site_description', '""'),
    ('photographer_name', '""'),
    ('about_text', '""'),
    ('social_links', '[]'),
    ('accent_color', '"#f97316"');
