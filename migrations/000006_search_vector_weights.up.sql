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

CREATE TRIGGER post_tags_search_vector_trigger
AFTER INSERT OR DELETE ON post_tags
FOR EACH ROW EXECUTE FUNCTION post_tags_search_vector_update();

-- Backfill: fire the posts trigger for every existing row.
UPDATE posts SET updated_at = updated_at;
