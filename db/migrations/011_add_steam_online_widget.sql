DELETE FROM widgets
WHERE service_id = (SELECT id FROM services WHERE name = 'steam')
  AND name = 'friends_online';

UPDATE widgets
SET display_name = 'Steam friends',
    description = 'Display Steam friends online or search the status of a specific friend'
WHERE service_id = (SELECT id FROM services WHERE name = 'steam')
  AND name = 'friends';

UPDATE widget_params
SET is_required = FALSE
WHERE widget_id = (
    SELECT w.id FROM widgets w
    JOIN services s ON s.id = w.service_id
    WHERE s.name = 'steam' AND w.name = 'friends'
)
AND name = 'pseudo';

INSERT INTO widget_params (widget_id, name, type, is_required, default_value)
SELECT w.id, params.name, params.type, FALSE, params.default_value
FROM widgets w
JOIN services s ON s.id = w.service_id
CROSS JOIN (VALUES
    ('mode', 'string', 'online'),
    ('limit', 'integer', '5'),
    ('inGameOnly', 'integer', '0')
) AS params(name, type, default_value)
WHERE s.name = 'steam' AND w.name = 'friends'
ON CONFLICT (widget_id, name) DO UPDATE SET
    type = EXCLUDED.type,
    is_required = EXCLUDED.is_required,
    default_value = EXCLUDED.default_value;

INSERT INTO widget_allowed_sizes (widget_id, size_code)
SELECT w.id, sizes.code
FROM widgets w
JOIN services s ON s.id = w.service_id
CROSS JOIN widget_sizes sizes
WHERE s.name = 'steam' AND w.name = 'friends'
ON CONFLICT (widget_id, size_code) DO NOTHING;
