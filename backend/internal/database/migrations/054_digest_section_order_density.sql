-- Custom digest templates: section order + density beyond include toggles.
ALTER TABLE user_preference
    ADD COLUMN IF NOT EXISTS digest_section_order JSONB NOT NULL DEFAULT '["highlights","topics","emails"]'::jsonb,
    ADD COLUMN IF NOT EXISTS digest_density TEXT NOT NULL DEFAULT 'comfortable';
