INSERT INTO widgets (service_id, name, display_name, description)
SELECT id, 'friends', 'Steam friend status', 'Display the Steam status of a friend searched by pseudo'
FROM services
WHERE name = 'steam'
ON CONFLICT (service_id, name) DO NOTHING;

INSERT INTO widget_params (widget_id, name, type, is_required)
SELECT w.id, 'pseudo', 'string', TRUE
FROM widgets w
JOIN services s ON s.id = w.service_id
WHERE s.name = 'steam' AND w.name = 'friends'
ON CONFLICT (widget_id, name) DO NOTHING;

INSERT INTO widget_allowed_sizes (widget_id, size_code)
SELECT w.id, sizes.code
FROM widgets w
JOIN services s ON s.id = w.service_id
CROSS JOIN widget_sizes sizes
WHERE s.name = 'steam' AND w.name = 'friends'
ON CONFLICT (widget_id, size_code) DO NOTHING;
