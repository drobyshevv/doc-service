CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL,

    title VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,

    s3_key TEXT NOT NULL UNIQUE,

    is_public BOOLEAN NOT NULL DEFAULT FALSE,

    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_documents_owner_id ON documents(owner_id);
CREATE INDEX idx_documents_is_public ON documents(is_public);
CREATE INDEX idx_documents_created_at ON documents(created_at);