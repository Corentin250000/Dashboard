-- OAuth-backed services used by the service connection routes.
INSERT INTO services (name, display_name, description, auth_type)
VALUES
    ('spotify', 'Spotify', 'Spotify account and playback data', 'oauth2'),
    ('steam', 'Steam', 'Steam account and game data', 'oauth2')
ON CONFLICT (name) DO NOTHING;
