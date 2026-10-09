INSERT INTO widgets (service_id, name, display_name, description)
SELECT id, 'friends_online', 'Steam friends online', 'Display Steam friends currently online'
FROM services WHERE name = 'steam'
ON CONFLICT (service_id, name) DO NOTHING;

INSERT INTO widget_params (widget_id, name, type, is_required, default_value)
SELECT w.id, params.name, 'integer', FALSE, params.default_value
FROM widgets w
JOIN services s ON s.id = w.service_id
CROSS JOIN (VALUES ('limit', '5'), ('inGameOnly', '0')) AS params(name, default_value)
WHERE s.name = 'steam' AND w.name = 'friends_online'
ON CONFLICT (widget_id, name) DO NOTHING;

INSERT INTO widget_allowed_sizes (widget_id, size_code)
SELECT w.id, sizes.code
FROM widgets w JOIN services s ON s.id = w.service_id CROSS JOIN widget_sizes sizes
WHERE s.name = 'steam' AND w.name = 'friends_online'
ON CONFLICT (widget_id, size_code) DO NOTHING;
