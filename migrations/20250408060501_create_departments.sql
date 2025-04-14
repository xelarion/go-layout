-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS departments
(
    id          BIGSERIAL PRIMARY KEY,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    name        VARCHAR(100)             NOT NULL,
    description TEXT                     NOT NULL DEFAULT '',
    enabled     BOOLEAN                  NOT NULL DEFAULT TRUE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_departments_name ON departments (name);

COMMENT ON TABLE departments IS 'Departments table';
COMMENT ON COLUMN departments.id IS 'Primary key';
COMMENT ON COLUMN departments.created_at IS 'Creation timestamp';
COMMENT ON COLUMN departments.name IS 'Department name';
COMMENT ON COLUMN departments.description IS 'Department description';
COMMENT ON COLUMN departments.enabled IS 'Department enabled status';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS departments;
-- +goose StatementEnd
