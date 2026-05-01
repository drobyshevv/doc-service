CREATE TABLE autocomplete_terms (
    id BIGSERIAL PRIMARY KEY,
    term TEXT NOT NULL UNIQUE,
    frequency INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_autocomplete_term ON autocomplete_terms(term);