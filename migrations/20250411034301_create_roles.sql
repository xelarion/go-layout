-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS roles
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      BIGINT                   NOT NULL DEFAULT 0,
    created_by      BIGINT                   NOT NULL DEFAULT 0,
    updated_by      BIGINT                   NOT NULL DEFAULT 0,
    created_role_id BIGINT                   NOT NULL DEFAULT 0,
    name            VARCHAR(100)             NOT NULL,
    slug            VARCHAR(100)             NOT NULL,
    code            VARCHAR(100)             NOT NULL,
    display_order   INT                      NOT NULL DEFAULT 1,
    description     TEXT                     NOT NULL DEFAULT '',
    enabled         BOOLEAN                  NOT NULL DEFAULT TRUE,
    permissions     VARCHAR(100)[]           NOT NULL,
    data_scopes     JSONB                    NOT NULL DEFAULT '[]'
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_name ON roles (name) WHERE deleted_at = 0;
CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_code ON roles (code) WHERE deleted_at = 0;

COMMENT ON TABLE roles IS '角色表';
COMMENT ON COLUMN roles.id IS '主键ID';
COMMENT ON COLUMN roles.created_at IS '创建时间';
COMMENT ON COLUMN roles.updated_at IS '更新时间';
COMMENT ON COLUMN roles.deleted_at IS '删除时间';
COMMENT ON COLUMN roles.created_by IS '创建人ID';
COMMENT ON COLUMN roles.updated_by IS '更新人ID';
COMMENT ON COLUMN roles.created_role_id IS '创建时角色ID';
COMMENT ON COLUMN roles.name IS '名称';
COMMENT ON COLUMN roles.slug IS '标识';
COMMENT ON COLUMN roles.code IS '编号';
COMMENT ON COLUMN roles.display_order IS '显示顺序';
COMMENT ON COLUMN roles.description IS '描述';
COMMENT ON COLUMN roles.enabled IS '启用状态';
COMMENT ON COLUMN roles.permissions IS '权限';
COMMENT ON COLUMN roles.data_scopes IS '数据范围';

INSERT INTO roles (name, slug, code, description, enabled, permissions)
VALUES ('Super Admin', 'super_admin', 'JS000', 'Super admin role', TRUE, '{*}')
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
