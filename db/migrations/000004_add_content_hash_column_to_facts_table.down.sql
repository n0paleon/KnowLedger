DROP INDEX IF EXISTS idx_facts_content_hash;
ALTER TABLE facts DROP COLUMN IF EXISTS content_hash;
