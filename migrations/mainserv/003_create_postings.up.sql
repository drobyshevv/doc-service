CREATE TABLE postings (
    id BIGSERIAL PRIMARY KEY,

    term_id BIGINT NOT NULL REFERENCES terms(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,

    term_frequency INTEGER NOT NULL,
    tfidf_score DOUBLE PRECISION NOT NULL DEFAULT 0,

    UNIQUE(term_id, document_id)
);

CREATE INDEX idx_postings_term_id ON postings(term_id);
CREATE INDEX idx_postings_document_id ON postings(document_id);
CREATE INDEX idx_postings_tfidf_score ON postings(tfidf_score DESC);