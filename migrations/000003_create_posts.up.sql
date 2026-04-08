CREATE TYPE post_status AS ENUM ('draft', 'published');

CREATE TABLE tags (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_tags_slug ON tags (slug);

CREATE TABLE posts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    content         TEXT NOT NULL DEFAULT '',
    excerpt         TEXT NOT NULL DEFAULT '',
    cover_image_url TEXT NOT NULL DEFAULT '',
    status          post_status NOT NULL DEFAULT 'draft',
    author_id       UUID NOT NULL REFERENCES users(id),
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ,
    search_vector   TSVECTOR,
    view_count      BIGINT NOT NULL DEFAULT 0,
    comment_count   INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_posts_slug       ON posts (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_author_id  ON posts (author_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_status     ON posts (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_deleted_at ON posts (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_posts_search     ON posts USING GIN (search_vector);

-- Full-text search trigger
CREATE OR REPLACE FUNCTION posts_search_vector_update() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', NEW.title || ' ' || NEW.content);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER posts_search_vector_trigger
BEFORE INSERT OR UPDATE ON posts
FOR EACH ROW EXECUTE FUNCTION posts_search_vector_update();

CREATE TABLE post_tags (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);

CREATE TYPE comment_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id    UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id  UUID NOT NULL REFERENCES users(id),
    body       TEXT NOT NULL,
    status     comment_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_comments_post_id    ON comments (post_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_author_id  ON comments (author_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_comments_deleted_at ON comments (deleted_at) WHERE deleted_at IS NOT NULL;

-- Trigger: maintain comment_count on posts
CREATE OR REPLACE FUNCTION update_post_comment_count() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' AND NEW.status = 'approved' AND NEW.deleted_at IS NULL THEN
        UPDATE posts SET comment_count = comment_count + 1 WHERE id = NEW.post_id;
    ELSIF TG_OP = 'UPDATE' THEN
        IF OLD.status != 'approved' AND NEW.status = 'approved' AND NEW.deleted_at IS NULL THEN
            UPDATE posts SET comment_count = comment_count + 1 WHERE id = NEW.post_id;
        ELSIF OLD.status = 'approved' AND (NEW.status != 'approved' OR NEW.deleted_at IS NOT NULL) THEN
            UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id = NEW.post_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER comment_count_trigger
AFTER INSERT OR UPDATE ON comments
FOR EACH ROW EXECUTE FUNCTION update_post_comment_count();

CREATE TABLE images (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename    TEXT NOT NULL,
    s3_key      TEXT NOT NULL UNIQUE,
    url         TEXT NOT NULL,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
