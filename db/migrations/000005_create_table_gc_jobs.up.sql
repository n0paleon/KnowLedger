CREATE TYPE gc_job_status AS ENUM ('pending', 'running', 'completed', 'failed');
CREATE TYPE gc_job_trigger AS ENUM ('automatic', 'manual');

CREATE TABLE gc_jobs (
                         id           VARCHAR(26) PRIMARY KEY,
                         status       gc_job_status NOT NULL DEFAULT 'pending',
                         trigger      gc_job_trigger NOT NULL,
                         started_at   TIMESTAMP,
                         finished_at  TIMESTAMP,
                         created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE gc_job_logs (
                             id         BIGSERIAL PRIMARY KEY,
                             job_id     VARCHAR(26) NOT NULL REFERENCES gc_jobs(id) ON DELETE CASCADE,
                             level      VARCHAR(10) NOT NULL,   -- 'info' | 'error' | 'debug'
                             message    TEXT NOT NULL,
                             created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gc_job_logs_job_id ON gc_job_logs(job_id);