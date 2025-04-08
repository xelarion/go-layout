-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id         BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    username   VARCHAR(100)             NOT NULL,
    email      VARCHAR(100)             NOT NULL,
    password   VARCHAR(100)             NOT NULL,
    enabled    BOOLEAN                  NOT NULL DEFAULT TRUE,
    role       VARCHAR(20)              NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username);

INSERT INTO users (username, email, password, role)
VALUES ('admin',
        'admin@example.com',
        '$2a$10$VN1FxiFmk1gm4txt2VIJxOLNFKOhyFqm1.EQ2B4rH746u6fQwJOia',
        'admin')
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
