-- Remove entry_id from guilt_scores
DROP INDEX IF EXISTS idx_guilt_scores_entry_id;
ALTER TABLE guilt_scores DROP COLUMN IF EXISTS entry_id;
