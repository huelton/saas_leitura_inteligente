-- Fase 3: jobs assíncronos e uso diário (freemium)

CREATE TABLE IF NOT EXISTS processing_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    file_path TEXT NOT NULL,
    title TEXT NOT NULL,
    author TEXT,
    error_message TEXT,
    book_id UUID REFERENCES books (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_processing_jobs_user ON processing_jobs (user_id);
CREATE INDEX IF NOT EXISTS idx_processing_jobs_status ON processing_jobs (status);

CREATE TABLE IF NOT EXISTS usage_daily (
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    day_utc DATE NOT NULL,
    questions_generated INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, day_utc)
);
