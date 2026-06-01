DROP INDEX IF EXISTS idx_documents_filename_trgm;
DROP INDEX IF EXISTS idx_documents_title_trgm;
-- pg_trgm не удаляем — может использоваться другими частями системы