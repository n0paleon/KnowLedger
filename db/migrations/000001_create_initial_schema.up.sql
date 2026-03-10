-- ================================
-- Create ENUM type for fact status
-- ================================
DO $$
BEGIN
CREATE TYPE fact_status AS ENUM ('draft', 'published');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;


-- ================================
-- Facts table
-- ================================
CREATE TABLE IF NOT EXISTS facts (
     id VARCHAR(26) PRIMARY KEY, -- ULID length selalu 26
    content TEXT NOT NULL,

    -- status fact: draft / published
    status fact_status NOT NULL DEFAULT 'draft',

    -- optional source URL
    source_url TEXT,

    -- media attachments (images, videos, etc.)
    -- contoh value: [{"type": "image", "url": "https://..."}]
    media JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
                             );

-- Index penting untuk query public
-- karena hampir semua query akan filter status = 'published'
-- dan biasanya diurutkan berdasarkan created_at
CREATE INDEX IF NOT EXISTS idx_facts_status_created_at
    ON facts(status, created_at DESC);



-- ================================
-- Tags table
-- ================================
CREATE TABLE IF NOT EXISTS tags (
                                    id VARCHAR(26) PRIMARY KEY,

    -- nama tag harus unik
    name VARCHAR(100) NOT NULL UNIQUE
    );

-- membuat tag case-insensitive
-- jadi 'Science', 'science', 'SCIENCE' dianggap sama
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_tag_name_lower
    ON tags (LOWER(name));



-- ================================
-- Join table: facts <-> tags
-- ================================
CREATE TABLE IF NOT EXISTS fact_tags (
                                         fact_id VARCHAR(26) NOT NULL,
    tag_id VARCHAR(26) NOT NULL,

    -- composite primary key untuk mencegah duplikasi
    PRIMARY KEY (fact_id, tag_id),

    CONSTRAINT fk_fact
    FOREIGN KEY (fact_id)
    REFERENCES facts(id)
    ON DELETE CASCADE,

    CONSTRAINT fk_tag
    FOREIGN KEY (tag_id)
    REFERENCES tags(id)
    ON DELETE CASCADE
    );

-- index untuk mempercepat query:
-- mencari semua facts dengan tag tertentu
-- contoh query:
-- SELECT * FROM facts
-- JOIN fact_tags ON ...
-- WHERE tag_id = ?
CREATE INDEX IF NOT EXISTS idx_fact_tags_tag_id
    ON fact_tags(tag_id);
