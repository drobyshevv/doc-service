CREATE TABLE terms (
    id BIGSERIAL PRIMARY KEY,
    term TEXT NOT NULL UNIQUE,
    document_frequency INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_terms_term ON terms(term);