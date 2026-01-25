-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_blobs
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    key        VARCHAR(255)             NOT NULL,
    mime_type  VARCHAR(255)             NOT NULL,
    filename   VARCHAR(255)             NOT NULL,
    byte_size  BIGINT                   NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_storage_blobs_key ON storage_blobs (key);

COMMENT ON TABLE storage_blobs IS '存储文件';
COMMENT ON COLUMN storage_blobs.id IS '主键';
COMMENT ON COLUMN storage_blobs.created_at IS '创建时间';
COMMENT ON COLUMN storage_blobs.key IS '文件唯一标识';
COMMENT ON COLUMN storage_blobs.mime_type IS '文件类型, 例如: image/jpeg';
COMMENT ON COLUMN storage_blobs.filename IS '文件名';
COMMENT ON COLUMN storage_blobs.byte_size IS '文件大小(字节)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS storage_blobs;
-- +goose StatementEnd
