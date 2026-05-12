-- Fase 4: otimização de custo de IA com cache persistente de MCQ por página.

ALTER TABLE questions
    ADD COLUMN IF NOT EXISTS question_type TEXT NOT NULL DEFAULT 'open';

ALTER TABLE questions
    ADD COLUMN IF NOT EXISTS options JSONB;

ALTER TABLE questions
    ADD COLUMN IF NOT EXISTS correct_idx INT;

CREATE INDEX IF NOT EXISTS idx_questions_book_page_type
    ON questions (book_id, page_number, question_type);
