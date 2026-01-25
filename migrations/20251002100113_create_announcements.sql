-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS announcements
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      BIGINT                   NOT NULL DEFAULT 0,
    created_by      BIGINT                   NOT NULL DEFAULT 0,
    updated_by      BIGINT                   NOT NULL DEFAULT 0,
    created_role_id BIGINT                   NOT NULL DEFAULT 0,
    title           VARCHAR(255)             NOT NULL DEFAULT '',
    content         TEXT                     NOT NULL DEFAULT '',
    status          SMALLINT                 NOT NULL DEFAULT 0,
    published_at    TIMESTAMP WITH TIME ZONE NULL,
    publisher_id    BIGINT                   NULL
);

COMMENT ON TABLE announcements IS '公告';
COMMENT ON COLUMN announcements.id IS '主键';
COMMENT ON COLUMN announcements.created_at IS '创建时间';
COMMENT ON COLUMN announcements.updated_at IS '更新时间';
COMMENT ON COLUMN announcements.deleted_at IS '删除时间';
COMMENT ON COLUMN announcements.created_by IS '创建人';
COMMENT ON COLUMN announcements.updated_by IS '更新人';
COMMENT ON COLUMN announcements.created_role_id IS '创建时角色ID';
COMMENT ON COLUMN announcements.title IS '标题';
COMMENT ON COLUMN announcements.content IS '内容';
COMMENT ON COLUMN announcements.status IS '状态';
COMMENT ON COLUMN announcements.published_at IS '发布时间';
COMMENT ON COLUMN announcements.publisher_id IS '发布人';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS announcements;
-- +goose StatementEnd
