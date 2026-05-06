DROP INDEX IF EXISTS idx_documents_token_count;

ALTER TABLE documents
DROP COLUMN IF EXISTS token_count;