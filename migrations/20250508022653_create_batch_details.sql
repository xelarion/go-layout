-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS batch_details
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    batch_record_id BIGINT                   NOT NULL,
    key             VARCHAR(100)             NOT NULL,
    data            JSONB                    NOT NULL DEFAULT '{}',
    status          SMALLINT                 NOT NULL DEFAULT 0,
    failure_reason  TEXT                     NOT NULL DEFAULT ''
);

CREATE INDEX idx_batch_details_batch_record_id ON batch_details (batch_record_id);

COMMENT ON TABLE batch_details IS '批量处理明细';
COMMENT ON COLUMN batch_details.id IS '主键';
COMMENT ON COLUMN batch_details.created_at IS '创建时间';
COMMENT ON COLUMN batch_details.batch_record_id IS '批量处理记录ID';
COMMENT ON COLUMN batch_details.key IS '关键字段(用于标识此行数据或搜索)';
COMMENT ON COLUMN batch_details.data IS '数据';
COMMENT ON COLUMN batch_details.status IS '状态';
COMMENT ON COLUMN batch_details.failure_reason IS '失败原因';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS batch_details;
-- +goose StatementEnd
