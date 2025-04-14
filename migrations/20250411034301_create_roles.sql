-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS roles
(
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    name        VARCHAR(100)             NOT NULL,
    slug        VARCHAR(100)             NOT NULL,
    description TEXT                     NOT NULL DEFAULT '',
    enabled     BOOLEAN                  NOT NULL DEFAULT TRUE,
    permissions VARCHAR(100)[]           NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name ON roles (name);

COMMENT ON TABLE roles IS 'Roles table';
COMMENT ON COLUMN roles.id IS 'Primary key';
COMMENT ON COLUMN roles.created_at IS 'Creation timestamp';
COMMENT ON COLUMN roles.name IS 'Role name';
COMMENT ON COLUMN roles.slug IS 'Role slug';
COMMENT ON COLUMN roles.description IS 'Role description';
COMMENT ON COLUMN roles.enabled IS 'Role enabled status';
COMMENT ON COLUMN roles.permissions IS 'Role permissions';

INSERT INTO roles (name, slug, description, enabled, permissions)
VALUES ('Super Admin', 'super_admin', 'Super admin role', TRUE, '{}')
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
