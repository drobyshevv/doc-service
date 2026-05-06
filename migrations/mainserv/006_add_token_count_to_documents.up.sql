ALTER TABLE documents
ADD COLUMN token_count INTEGER NOT NULL DEFAULT 0;

CREATE INDEX idx_documents_token_count ON documents(token_count);