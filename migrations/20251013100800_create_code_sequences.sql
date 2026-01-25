-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS code_sequences
(
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    rule_slug      VARCHAR(100)             NOT NULL,
    time_key       VARCHAR(50)              NOT NULL,
    current_serial BIGINT                   NOT NULL DEFAULT 0,
    PRIMARY KEY (rule_slug, time_key)
);

COMMENT ON TABLE code_sequences IS '编码流水号序列表';
COMMENT ON COLUMN code_sequences.created_at IS '创建时间';
COMMENT ON COLUMN code_sequences.rule_slug IS '编码规则标识（主键1）';
COMMENT ON COLUMN code_sequences.time_key IS '时间分组键（主键2）';
COMMENT ON COLUMN code_sequences.current_serial IS '当前流水号';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS code_sequences;
-- +goose StatementEnd
