-- Enable pg_trgm extension
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Add search_vector column to posts
ALTER TABLE posts ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Create update function for post search vector
CREATE OR REPLACE FUNCTION update_post_search_vector() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(NEW.body, '')), 'B');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop trigger if exists, then create
DROP TRIGGER IF EXISTS trg_post_search_vector ON posts;
CREATE TRIGGER trg_post_search_vector
  BEFORE INSERT OR UPDATE ON posts
  FOR EACH ROW EXECUTE FUNCTION update_post_search_vector();

-- GIN index on search_vector
CREATE INDEX IF NOT EXISTS idx_posts_search ON posts USING GIN(search_vector);

-- Trigram indexes for users and communities
CREATE INDEX IF NOT EXISTS idx_users_username_trgm ON users USING GIN(username gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_display_name_trgm ON users USING GIN(display_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_communities_name_trgm ON communities USING GIN(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_communities_slug_trgm ON communities USING GIN(slug gin_trgm_ops);
