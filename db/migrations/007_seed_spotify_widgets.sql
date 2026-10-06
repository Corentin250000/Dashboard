-- Spotify widget catalog. This migration is idempotent.
INSERT INTO services (name, display_name, description, auth_type)
VALUES ('spotify', 'Spotify', 'Spotify account and playback data', 'oauth2')
ON CONFLICT (name) DO NOTHING;
INSERT INTO widgets (service_id, name, display_name, description)
SELECT
    id,
    'currently_playing',
    'Currently playing',
    'Display the track currently playing on Spotify'
FROM services
WHERE name = 'spotify'
ON CONFLICT (service_id, name) DO NOTHING;

INSERT INTO widgets (service_id, name, display_name, description)
SELECT
    id,
    'top_items',
    'Top items',
    'Display the user''s top Spotify artists or tracks'
FROM services
WHERE name = 'spotify'
ON CONFLICT (service_id, name) DO NOTHING;

INSERT INTO widget_params (widget_id, name, type, is_required, default_value)
SELECT
    w.id,
    p.name,
    p.type,
    p.required,
    p.default_value
FROM widgets w
JOIN services s ON s.id = w.service_id
CROSS JOIN (
    VALUES
        ('layout', 'string', FALSE, 'compact'),
        ('show_progress', 'integer', FALSE, '1')
) AS p(name, type, required, default_value)
WHERE s.name = 'spotify'
  AND w.name = 'currently_playing'
ON CONFLICT (widget_id, name) DO NOTHING;

INSERT INTO widget_params (widget_id, name, type, is_required, default_value)
SELECT
    w.id,
    p.name,
    p.type,
    p.required,
    p.default_value
FROM widgets w
JOIN services s ON s.id = w.service_id
CROSS JOIN (
    VALUES
        ('item_type', 'string', TRUE, NULL),
        ('time_range', 'string', TRUE, NULL),
        ('limit', 'integer', FALSE, '5')
) AS p(name, type, required, default_value)
WHERE s.name = 'spotify'
  AND w.name = 'top_items'
ON CONFLICT (widget_id, name) DO NOTHING;

INSERT INTO widget_allowed_sizes (widget_id, size_code)
SELECT
    w.id,
    sizes.code
FROM widgets w
JOIN services s ON s.id = w.service_id
CROSS JOIN widget_sizes sizes
WHERE s.name = 'spotify'
  AND w.name IN ('currently_playing', 'top_items')
ON CONFLICT (widget_id, size_code) DO NOTHING;
