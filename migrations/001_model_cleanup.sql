-- Migration: Add new fields to llm_models table and migrate data from Settings
-- This migration is idempotent and can be run multiple times safely.

-- 1. Add new columns (SQLite compatible syntax)
-- For SQLite, we use ALTER TABLE ADD COLUMN which ignores duplicates if column exists
-- For MySQL/PostgreSQL, you may need to check column existence first

-- SQLite version (default):
ALTER TABLE llm_models ADD COLUMN generation_mode VARCHAR(64);
ALTER TABLE llm_models ADD COLUMN endpoint_path VARCHAR(255);
ALTER TABLE llm_models ADD COLUMN supports_streaming BOOLEAN DEFAULT 0;
ALTER TABLE llm_models ADD COLUMN supports_cancel BOOLEAN DEFAULT 0;

-- Note: The migration of Settings JSON data to the new columns is handled by GORM AutoMigrate
-- and application-level code in seed.go. The SQL below is for reference if manual migration is needed.

-- For MySQL/MariaDB:
-- UPDATE llm_models
-- SET generation_mode = JSON_UNQUOTE(JSON_EXTRACT(settings, '$.mode')),
--     endpoint_path = JSON_UNQUOTE(JSON_EXTRACT(settings, '$.endpoint'))
-- WHERE settings IS NOT NULL
--   AND JSON_EXTRACT(settings, '$.mode') IS NOT NULL;

-- For PostgreSQL:
-- UPDATE llm_models
-- SET generation_mode = settings->>'mode',
--     endpoint_path = settings->>'endpoint'
-- WHERE settings IS NOT NULL
--   AND settings->>'mode' IS NOT NULL;

-- For SQLite (json_extract):
-- UPDATE llm_models
-- SET generation_mode = json_extract(settings, '$.mode'),
--     endpoint_path = json_extract(settings, '$.endpoint')
-- WHERE settings IS NOT NULL
--   AND json_extract(settings, '$.mode') IS NOT NULL;
