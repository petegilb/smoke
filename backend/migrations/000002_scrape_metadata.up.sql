CREATE TABLE scrape_metadata (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO scrape_metadata (key, value) VALUES ('last_scraped_at', '1970-01-01T00:00:00Z');
