-- Voter Trust Score for Weighted Review Voting
-- When a user casts a vote on a moderation action, we store their current
-- trust_score so that EvaluateUnfairPenalty can weigh votes by trust.
ALTER TABLE mod_decision_reviews ADD COLUMN IF NOT EXISTS voter_trust_score REAL DEFAULT 1.0;
