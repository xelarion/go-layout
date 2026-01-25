-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notifications
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by BIGINT                   NOT NULL DEFAULT 0,
    title      VARCHAR(255)             NOT NULL DEFAULT '',
    content    TEXT                     NOT NULL DEFAULT '',
    type       SMALLINT                 NOT NULL DEFAULT 0,
    category   SMALLINT                 NOT NULL DEFAULT 0,
    scope      SMALLINT                 NOT NULL DEFAULT 0,
    payload    JSONB                    NOT NULL DEFAULT '{}'
);

COMMENT ON TABLE notifications IS '通知';
COMMENT ON COLUMN notifications.id IS '主键';
COMMENT ON COLUMN notifications.created_at IS '创建时间';
COMMENT ON COLUMN notifications.created_by IS '创建人';
COMMENT ON COLUMN notifications.title IS '标题';
COMMENT ON COLUMN notifications.content IS '内容';
COMMENT ON COLUMN notifications.type IS '类型';
COMMENT ON COLUMN notifications.category IS '分类';
COMMENT ON COLUMN notifications.scope IS '范围';
COMMENT ON COLUMN notifications.payload IS '数据';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd
