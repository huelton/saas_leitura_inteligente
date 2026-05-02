-- Fase 2: chapter_summaries + flashcards (mesmo conteúdo de internal/database/schema_phase2.sql)

CREATE TABLE IF NOT EXISTS chapter_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id UUID NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    chapter_number INT NOT NULL,
    page_from INT NOT NULL,
    page_to INT NOT NULL,
    summary TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (book_id, chapter_number)
);

CREATE TABLE IF NOT EXISTS flashcards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    book_id UUID NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    page_number INT,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    interval_days INT NOT NULL DEFAULT 1,
    ease_factor DOUBLE PRECISION NOT NULL DEFAULT 2.5,
    next_review_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flashcards_user_due ON flashcards (user_id, next_review_date);

CREATE INDEX IF NOT EXISTS idx_chapter_summaries_book ON chapter_summaries (book_id);
