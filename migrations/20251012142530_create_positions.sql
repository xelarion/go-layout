-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS positions
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
    organization_id BIGINT                   NOT NULL,
    remark          TEXT                     NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_positions_name ON positions (name) WHERE deleted_at = 0;
CREATE UNIQUE INDEX IF NOT EXISTS idx_positions_code ON positions (code) WHERE deleted_at = 0;


COMMENT ON TABLE positions IS '岗位表';
COMMENT ON COLUMN positions.id IS '主键';
COMMENT ON COLUMN positions.created_at IS '创建时间';
COMMENT ON COLUMN positions.updated_at IS '更新时间';
COMMENT ON COLUMN positions.deleted_at IS '删除时间';
COMMENT ON COLUMN positions.created_by IS '创建人ID';
COMMENT ON COLUMN positions.updated_by IS '更新人ID';
COMMENT ON COLUMN positions.created_role_id IS '创建时角色ID';
COMMENT ON COLUMN positions.name IS '岗位名称';
COMMENT ON COLUMN positions.code IS '岗位编号';
COMMENT ON COLUMN positions.organization_id IS '所属组织';
COMMENT ON COLUMN positions.remark IS '备注';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS positions;
-- +goose StatementEnd
