-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id            BIGSERIAL PRIMARY KEY,
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    username      VARCHAR(100)             NOT NULL,
    password      VARCHAR(255)             NOT NULL,
    full_name     VARCHAR(100)             NOT NULL,
    phone_number  VARCHAR(20)              NOT NULL,
    email         VARCHAR(100)             NOT NULL,
    enabled       BOOLEAN                  NOT NULL DEFAULT TRUE,
    department_id BIGINT                   NOT NULL,
    role_id       BIGINT                   NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username);

COMMENT ON TABLE users IS 'Users table';
COMMENT ON COLUMN users.id IS 'Primary key';
COMMENT ON COLUMN users.created_at IS 'Creation timestamp';
COMMENT ON COLUMN users.username IS 'Username';
COMMENT ON COLUMN users.password IS 'Password';
COMMENT ON COLUMN users.full_name IS 'Full name';
COMMENT ON COLUMN users.phone_number IS 'Phone number';
COMMENT ON COLUMN users.email IS 'Email';
COMMENT ON COLUMN users.enabled IS 'Enabled';
COMMENT ON COLUMN users.role_id IS 'Role ID';
COMMENT ON COLUMN users.department_id IS 'Department ID';

INSERT INTO users (username, password, full_name, phone_number, email, enabled, department_id, role_id)
VALUES ('admin',
        '$2a$10$VN1FxiFmk1gm4txt2VIJxOLNFKOhyFqm1.EQ2B4rH746u6fQwJOia',
        'Admin',
        '12345678900',
        'admin@example.com',
        TRUE,
        1,
        1)
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
