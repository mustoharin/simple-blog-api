CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT ''
);

CREATE TABLE permissions (
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Seed roles
INSERT INTO roles (id, name, description) VALUES
    (gen_random_uuid(), 'superadmin', 'Full system access'),
    (gen_random_uuid(), 'admin',      'Administrative access'),
    (gen_random_uuid(), 'editor',     'Content management'),
    (gen_random_uuid(), 'commenter',  'Comment only');

-- Seed permissions
INSERT INTO permissions (id, name) VALUES
    (gen_random_uuid(), 'post:create'),
    (gen_random_uuid(), 'post:edit'),
    (gen_random_uuid(), 'post:delete'),
    (gen_random_uuid(), 'post:publish'),
    (gen_random_uuid(), 'image:upload'),
    (gen_random_uuid(), 'comment:create'),
    (gen_random_uuid(), 'comment:approve'),
    (gen_random_uuid(), 'user:manage'),
    (gen_random_uuid(), 'user:create'),
    (gen_random_uuid(), 'user:update'),
    (gen_random_uuid(), 'user:delete'),
    (gen_random_uuid(), 'audit:read'),
    (gen_random_uuid(), 'dashboard:read');

-- Assign permissions to roles
-- superadmin: all
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p WHERE r.name = 'superadmin';

-- admin: all except user:create, user:update, user:delete (managed separately)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON p.name IN (
    'post:create','post:edit','post:delete','post:publish',
    'image:upload','comment:create','comment:approve',
    'user:manage','audit:read','dashboard:read'
) WHERE r.name = 'admin';

-- editor: post ops, image upload, comment create/approve, dashboard
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON p.name IN (
    'post:create','post:edit','post:delete','post:publish',
    'image:upload','comment:create','comment:approve','dashboard:read'
) WHERE r.name = 'editor';

-- commenter: comment:create only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p ON p.name = 'comment:create'
WHERE r.name = 'commenter';
