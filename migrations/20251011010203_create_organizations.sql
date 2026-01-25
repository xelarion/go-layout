-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS organizations
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      BIGINT                   NOT NULL DEFAULT 0,
    created_by      BIGINT                   NOT NULL DEFAULT 0,
    updated_by      BIGINT                   NOT NULL DEFAULT 0,
    created_role_id BIGINT                   NOT NULL DEFAULT 0,
    name            VARCHAR(100)             NOT NULL,
    code            VARCHAR(100)             NOT NULL,
    parent_id       BIGINT                   NULl,
    type            smallint                 NOT NULL,
    owner_id        BIGINT                   NOT NULL DEFAULT 0,
    remark          TEXT                     NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX idx_organizations_name ON organizations (name) WHERE deleted_at = 0;
CREATE UNIQUE INDEX idx_organizations_code ON organizations (code) WHERE deleted_at = 0;

COMMENT ON TABLE organizations IS '组织机构表';
COMMENT ON COLUMN organizations.id IS '主键ID';
COMMENT ON COLUMN organizations.created_at IS '创建时间';
COMMENT ON COLUMN organizations.updated_at IS '更新时间';
COMMENT ON COLUMN organizations.deleted_at IS '删除时间';
COMMENT ON COLUMN organizations.created_by IS '创建人ID';
COMMENT ON COLUMN organizations.updated_by IS '更新人ID';
COMMENT ON COLUMN organizations.created_role_id IS '创建时角色ID';
COMMENT ON COLUMN organizations.name IS '组织名称';
COMMENT ON COLUMN organizations.code IS '组织编码';
COMMENT ON COLUMN organizations.parent_id IS '上级组织ID';
COMMENT ON COLUMN organizations.type IS '组织类型';
COMMENT ON COLUMN organizations.owner_id IS '负责人ID';
COMMENT ON COLUMN organizations.remark IS '备注';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS organizations;
-- +goose StatementEnd