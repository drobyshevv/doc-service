CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE document_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    s3_key TEXT NOT NULL UNIQUE,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    uploaded_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    note TEXT,
    UNIQUE(document_id, version)
);

CREATE INDEX idx_versions_doc_id ON document_versions(document_id);
CREATE INDEX idx_versions_created_at ON document_versions(created_at DESC);

ALTER TABLE documents ADD COLUMN current_version INTEGER NOT NULL DEFAULT 1;
CREATE INDEX idx_documents_current_version ON documents(current_version);