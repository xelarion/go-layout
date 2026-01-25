-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_notifications
(
    id              BIGSERIAL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    notification_id BIGINT                   NOT NULL,
    user_id         BIGINT                   NOT NULL,
    is_read         BOOLEAN                  NOT NULL DEFAULT FALSE,
    read_at         TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX IF NOT EXISTS idx_user_notifications_user_created ON user_notifications (user_id, created_at DESC);

COMMENT ON TABLE user_notifications IS '用户通知';
COMMENT ON COLUMN user_notifications.id IS '主键';
COMMENT ON COLUMN user_notifications.created_at IS '创建时间';
COMMENT ON COLUMN user_notifications.notification_id IS '通知ID';
COMMENT ON COLUMN user_notifications.user_id IS '用户ID';
COMMENT ON COLUMN user_notifications.is_read IS '是否已读';
COMMENT ON COLUMN user_notifications.read_at IS '读取时间';


-- Create initial partition for current month and next month
DO
$$
    DECLARE
        current_month_start TIMESTAMP WITH TIME ZONE;
        next_month_start    TIMESTAMP WITH TIME ZONE;
        month_after_next    TIMESTAMP WITH TIME ZONE;
    BEGIN
        -- Get the start of current month in UTC
        current_month_start := date_trunc('month', CURRENT_TIMESTAMP AT TIME ZONE 'UTC');
        next_month_start := current_month_start + interval '1 month';
        month_after_next := next_month_start + interval '1 month';

        -- Create partition for current month
        EXECUTE format(
                'CREATE TABLE user_notifications_%s PARTITION OF user_notifications FOR VALUES FROM (%L) TO (%L)',
                to_char(current_month_start, 'YYYYMM'),
                current_month_start,
                next_month_start
                );

        -- Create partition for next month
        EXECUTE format(
                'CREATE TABLE user_notifications_%s PARTITION OF user_notifications FOR VALUES FROM (%L) TO (%L)',
                to_char(next_month_start, 'YYYYMM'),
                next_month_start,
                month_after_next
                );
    END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_notifications CASCADE;
-- +goose StatementEnd
