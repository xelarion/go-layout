-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS change_logs
(
    id            BIGSERIAL PRIMARY KEY,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    operator_id   BIGINT                   NOT NULL DEFAULT 0,
    resource_type VARCHAR(255)             NOT NULL,
    resource_id   BIGINT                   NOT NULL,
    action        SMALLINT                 NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_change_logs_resource_type_resource_id ON change_logs (resource_type, resource_id);

COMMENT ON TABLE change_logs IS '变更记录表';
COMMENT ON COLUMN change_logs.id IS '主键ID';
COMMENT ON COLUMN change_logs.created_at IS '创建时间';
COMMENT ON COLUMN change_logs.operator_id IS '操作人ID';
COMMENT ON COLUMN change_logs.resource_type IS '资源类型';
COMMENT ON COLUMN change_logs.resource_id IS '资源ID';
COMMENT ON COLUMN change_logs.action IS '操作类型';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS change_logs;
-- +goose StatementEnd
