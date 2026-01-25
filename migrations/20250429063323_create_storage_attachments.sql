-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_attachments
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      BIGINT                   NOT NULL DEFAULT 0,
    resource_type   VARCHAR(255)             NOT NULL,
    resource_id     BIGINT                   NOT NULL,
    storage_blob_id BIGINT                   NOT NULL,
    name            VARCHAR(255)             NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_storage_attachments_resource_type_resource_id ON storage_attachments (resource_type, resource_id) WHERE deleted_at = 0;
CREATE INDEX IF NOT EXISTS idx_storage_attachments_storage_blob_id ON storage_attachments (storage_blob_id) WHERE deleted_at = 0;

COMMENT ON TABLE storage_attachments IS '通用文件上传的附件信息';
COMMENT ON COLUMN storage_attachments.id IS '主键';
COMMENT ON COLUMN storage_attachments.created_at IS '创建时间';
COMMENT ON COLUMN storage_attachments.deleted_at IS '删除时间';
COMMENT ON COLUMN storage_attachments.resource_type IS '资源类型';
COMMENT ON COLUMN storage_attachments.resource_id IS '资源ID';
COMMENT ON COLUMN storage_attachments.storage_blob_id IS '存储文件ID';
COMMENT ON COLUMN storage_attachments.name IS '名称, 如: file, avatar (区分同一条记录不同类型的附件)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS storage_attachments;
-- +goose StatementEnd
