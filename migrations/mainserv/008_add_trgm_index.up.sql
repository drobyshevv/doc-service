-- Включаем расширение для триграммного поиска
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Индекс для быстрого поиска по названию
CREATE INDEX IF NOT EXISTS idx_documents_title_trgm 
ON documents USING GIN (title gin_trgm_ops);

-- Индекс для быстрого поиска по имени файла
CREATE INDEX IF NOT EXISTS idx_documents_filename_trgm 
ON documents USING GIN (original_filename gin_trgm_ops);