-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS change_log_details
(
    id            BIGSERIAL PRIMARY KEY,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    change_log_id BIGINT                   NOT NULL,
    field_name    VARCHAR(255)             NOT NULL,
    field_label   VARCHAR(255)             NOT NULL DEFAULT '',
    field_type    SMALLINT                 NOT NULL DEFAULT 1,
    relation_type SMALLINT                 NOT NULL DEFAULT 0,
    relation_meta JSONB                    NOT NULL DEFAULT '{}',
    old_value     TEXT                     NOT NULL,
    new_value     TEXT                     NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_change_log_details_change_log_id ON change_log_details (change_log_id);

COMMENT ON TABLE change_log_details IS '变更记录明细表';
COMMENT ON COLUMN change_log_details.id IS '主键ID';
COMMENT ON COLUMN change_log_details.created_at IS '创建时间';
COMMENT ON COLUMN change_log_details.change_log_id IS '变更记录ID';
COMMENT ON COLUMN change_log_details.field_name IS '字段名(结构体字段名)';
COMMENT ON COLUMN change_log_details.field_label IS '字段显示名称';
COMMENT ON COLUMN change_log_details.field_type IS '字段类型: 1=普通字段, 2=组合字段, 3=关联字段';
COMMENT ON COLUMN change_log_details.relation_type IS '关联类型: 0=非关联, 1=一对一, 2=一对多';
COMMENT ON COLUMN change_log_details.relation_meta IS '关联元数据(JSON对象)';
COMMENT ON COLUMN change_log_details.old_value IS '变更前的值';
COMMENT ON COLUMN change_log_details.new_value IS '变更后的值';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS change_log_details;
-- +goose StatementEnd
