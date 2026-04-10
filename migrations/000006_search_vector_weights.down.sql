DROP TRIGGER IF EXISTS post_tags_search_vector_trigger ON post_tags;
DROP FUNCTION IF EXISTS post_tags_search_vector_update();

-- Restore the original simple trigger function.
CREATE OR REPLACE FUNCTION posts_search_vector_update() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', NEW.title || ' ' || NEW.content);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Backfill with simple vector.
UPDATE posts SET updated_at = updated_at;
