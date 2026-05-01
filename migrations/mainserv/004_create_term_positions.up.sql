CREATE TABLE term_positions (
    id BIGSERIAL PRIMARY KEY,

    posting_id BIGINT NOT NULL REFERENCES postings(id) ON DELETE CASCADE,
    position INTEGER NOT NULL
);

CREATE INDEX idx_term_positions_posting_id ON term_positions(posting_id);
CREATE INDEX idx_term_positions_position ON term_positions(position);