DROP INDEX IF EXISTS idx_documents_current_version;
ALTER TABLE documents DROP COLUMN IF EXISTS current_version;

DROP TABLE IF EXISTS document_versions;