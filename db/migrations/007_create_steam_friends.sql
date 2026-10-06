CREATE TABLE steam_friends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_steam_id TEXT NOT NULL,
    friend_steam_id TEXT NOT NULL,
    persona_name TEXT NOT NULL,
    search_name TEXT NOT NULL,
    profile_visibility SMALLINT NOT NULL,
    profile_configured BOOLEAN NOT NULL,
    persona_state SMALLINT NOT NULL,
    current_game_id TEXT,
    current_game_name TEXT,
    last_logoff_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_steam_friends_owner_friend UNIQUE (owner_steam_id, friend_steam_id)
);

CREATE INDEX idx_steam_friends_owner_steam_id ON steam_friends (owner_steam_id);
CREATE INDEX idx_steam_friends_search_name ON steam_friends (search_name);
CREATE INDEX idx_steam_friends_synced_at ON steam_friends (synced_at);
