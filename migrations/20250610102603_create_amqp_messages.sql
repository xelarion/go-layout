-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS amqp_messages
(
    id           BIGSERIAL PRIMARY KEY,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    exchange     VARCHAR(255)             NOT NULL,
    routing_keys VARCHAR(255)[]           NOT NULL,
    data         JSONB                    NOT NULL,
    options      JSONB                    NOT NULL DEFAULT '{}'
);


COMMENT ON TABLE amqp_messages IS '队列消息';
COMMENT ON COLUMN amqp_messages.id IS '主键';
COMMENT ON COLUMN amqp_messages.created_at IS '创建时间';
COMMENT ON COLUMN amqp_messages.exchange IS '交换机';
COMMENT ON COLUMN amqp_messages.routing_keys IS '路由键';
COMMENT ON COLUMN amqp_messages.data IS '数据';
COMMENT ON COLUMN amqp_messages.options IS '选项';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS amqp_messages;
-- +goose StatementEnd
