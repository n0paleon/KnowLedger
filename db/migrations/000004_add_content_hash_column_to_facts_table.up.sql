CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE facts ADD COLUMN IF NOT EXISTS content_hash VARCHAR(64);

UPDATE facts
SET content_hash = ENCODE(DIGEST(LOWER(TRIM(content)), 'sha256'), 'hex')
WHERE content_hash IS NULL;

DELETE FROM facts
WHERE id IN (
    SELECT id FROM (
                       SELECT id,
                              ROW_NUMBER() OVER (
                   PARTITION BY content_hash
                   ORDER BY id ASC
               ) AS rn
                       FROM facts
                   ) ranked
    WHERE rn > 1
);

ALTER TABLE facts ALTER COLUMN content_hash SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_facts_content_hash ON facts(content_hash);
