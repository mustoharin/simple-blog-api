-- Replace the simple trigger with a weighted one that includes excerpt and tags.

CREATE OR REPLACE FUNCTION posts_search_vector_update() RETURNS TRIGGER AS $$
DECLARE
    tag_names TEXT;
BEGIN
    SELECT string_agg(t.name, ' ')
    INTO tag_names
    FROM post_tags pt
    JOIN tags t ON pt.tag_id = t.id
    WHERE pt.post_id = NEW.id;

    NEW.search_vector :=
        setweight(to_tsvector('english', coalesce(NEW.title,   '')), 'A') ||
        setweight(to_tsvector('english', coalesce(NEW.excerpt, '')), 'B') ||
        setweight(to_tsvector('english', coalesce(NEW.content, '')), 'C') ||
        setweight(to_tsvector('english', coalesce(tag_names,   '')), 'D');

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger on post_tags: refresh search_vector when tags are added/removed.
-- We touch updated_at to fire the posts BEFORE trigger without changing data.
CREATE OR REPLACE FUNCTION post_tags_search_vector_update() RETURNS TRIGGER AS $$
DECLARE
    affected_post_id UUID;
BEGIN
    IF TG_OP = 'DELETE' THEN
        affected_post_id := OLD.post_id;
    ELSE
        affected_post_id := NEW.post_id;
    END IF;

    UPDATE posts SET updated_at = updated_at WHERE id = affected_post_id;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- CREATE OR REPLACE TRIGGER requires PostgreSQL 14+ (this project uses PG 16).
CREATE OR REPLACE TRIGGER post_tags_search_vector_trigger
AFTER INSERT OR DELETE ON post_tags
FOR EACH ROW EXECUTE FUNCTION post_tags_search_vector_update();

-- Backfill: fire the posts BEFORE trigger for every existing row so search_vector
-- is rebuilt with the new weighted function.
-- NOTE: This locks all rows in posts for the duration of the migration. Acceptable
-- for a blog with a small posts table; on large tables run this in batches out-of-band.
UPDATE posts SET updated_at = updated_at;
