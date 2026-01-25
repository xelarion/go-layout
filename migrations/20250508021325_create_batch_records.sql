-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS batch_records
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      BIGINT                   NOT NULL DEFAULT 0,
    created_by      BIGINT                   NOT NULL DEFAULT 0,
    created_role_id BIGINT                   NOT NULL DEFAULT 0,
    operation_type  SMALLINT                 NOT NULL DEFAULT 0,
    data            JSONB                    NOT NULL DEFAULT '{}',
    status          SMALLINT                 NOT NULL DEFAULT 0,
    failure_reason  TEXT                     NOT NULL DEFAULT ''
);

COMMENT ON TABLE batch_records IS '批量处理记录';
COMMENT ON COLUMN batch_records.id IS '主键';
COMMENT ON COLUMN batch_records.created_at IS '创建时间';
COMMENT ON COLUMN batch_records.updated_at IS '更新时间';
COMMENT ON COLUMN batch_records.deleted_at IS '删除时间';
COMMENT ON COLUMN batch_records.created_by IS '创建人';
COMMENT ON COLUMN batch_records.created_role_id IS '创建时角色ID';
COMMENT ON COLUMN batch_records.operation_type IS '操作类型';
COMMENT ON COLUMN batch_records.data IS '数据';
COMMENT ON COLUMN batch_records.status IS '状态';
COMMENT ON COLUMN batch_records.failure_reason IS '失败原因';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS batch_records;
-- +goose StatementEnd
