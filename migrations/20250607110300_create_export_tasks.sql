-- +goose Up
-- +goose StatementBegin
CREATE TABLE export_tasks
(
    id            BIGSERIAL PRIMARY KEY,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at    BIGINT                   NOT NULL DEFAULT 0,
    created_by    BIGINT                   NOT NULL,
    task_type     VARCHAR(50)              NOT NULL,
    filename      VARCHAR(255)             NOT NULL,
    status        SMALLINT                 NOT NULL DEFAULT 0,
    is_async      BOOLEAN                  NOT NULL DEFAULT FALSE,
    total_count   INT                      NOT NULL DEFAULT 0,
    params        JSONB                    NOT NULL DEFAULT '{}',
    error_message TEXT                     NOT NULL DEFAULT '',
    started_at    TIMESTAMP WITH TIME ZONE NULL,
    completed_at  TIMESTAMP WITH TIME ZONE NULL
);

-- 创建索引
CREATE INDEX idx_export_tasks_created_by ON export_tasks (created_by) WHERE deleted_at = 0;

-- 添加表注释
COMMENT ON TABLE export_tasks IS '导出任务表';
COMMENT ON COLUMN export_tasks.id IS '主键';
COMMENT ON COLUMN export_tasks.created_at IS '创建时间';
COMMENT ON COLUMN export_tasks.updated_at IS '更新时间';
COMMENT ON COLUMN export_tasks.deleted_at IS '删除时间';
COMMENT ON COLUMN export_tasks.created_by IS '创建人ID';
COMMENT ON COLUMN export_tasks.task_type IS '任务类型';
COMMENT ON COLUMN export_tasks.filename IS '导出文件名';
COMMENT ON COLUMN export_tasks.status IS '任务状态';
COMMENT ON COLUMN export_tasks.is_async IS '是否异步';
COMMENT ON COLUMN export_tasks.total_count IS '总记录数';
COMMENT ON COLUMN export_tasks.params IS '导出参数';
COMMENT ON COLUMN export_tasks.error_message IS '错误信息';
COMMENT ON COLUMN export_tasks.started_at IS '开始时间';
COMMENT ON COLUMN export_tasks.completed_at IS '完成时间';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS export_tasks;
-- +goose StatementEnd
