-- Add entry_id to guilt_scores for per-entry scoring
ALTER TABLE guilt_scores ADD COLUMN entry_id UUID REFERENCES guilt_entries(id) ON DELETE CASCADE;
CREATE INDEX idx_guilt_scores_entry_id ON guilt_scores(entry_id);
