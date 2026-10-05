-- SNCF station catalog from the Navitia stop_areas endpoint.
CREATE TABLE sncf_stations (
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL
);
