-- Widget sizes and per-widget size allow-lists.
CREATE TABLE IF NOT EXISTS widget_sizes (
    code   VARCHAR(16) PRIMARY KEY,
    width  INTEGER NOT NULL CHECK (width > 0 AND width <= 6),
    height INTEGER NOT NULL CHECK (height > 0 AND height <= 6)
);

INSERT INTO widget_sizes (code, width, height)
VALUES
    ('small', 1, 1),
    ('wide', 2, 1),
    ('tall', 1, 2),
    ('big', 2, 2)
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS widget_allowed_sizes (
    widget_id UUID NOT NULL REFERENCES widgets (id) ON DELETE CASCADE,
    size_code VARCHAR(16) NOT NULL REFERENCES widget_sizes (code),
    PRIMARY KEY (widget_id, size_code)
);
CREATE INDEX IF NOT EXISTS idx_widget_allowed_sizes_size_code
    ON widget_allowed_sizes (size_code);

-- Existing widgets remain valid with the default small size. The catalog can
-- add or remove options for each widget afterwards.
INSERT INTO widget_allowed_sizes (widget_id, size_code)
SELECT id, 'small' FROM widgets
ON CONFLICT (widget_id, size_code) DO NOTHING;

ALTER TABLE widget_instances
    ADD COLUMN IF NOT EXISTS size_code VARCHAR(16) NOT NULL DEFAULT 'small';

ALTER TABLE widget_instances
    DROP CONSTRAINT IF EXISTS fk_widget_instance_allowed_size;

ALTER TABLE widget_instances
    ADD CONSTRAINT fk_widget_instance_allowed_size
    FOREIGN KEY (widget_id, size_code)
    REFERENCES widget_allowed_sizes (widget_id, size_code);
