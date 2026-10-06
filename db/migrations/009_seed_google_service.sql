-- Google service connection catalog entry.
INSERT INTO services (name, display_name, description, auth_type)
VALUES (
    'google',
    'Google',
    'Google account and calendar data',
    'oauth2'
)
ON CONFLICT (name) DO NOTHING;
