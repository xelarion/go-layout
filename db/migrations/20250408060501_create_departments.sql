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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS departments;
-- +goose StatementEnd
