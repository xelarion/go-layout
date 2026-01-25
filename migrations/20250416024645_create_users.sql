-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
    id              BIGSERIAL PRIMARY KEY,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at      BIGINT                   NOT NULL DEFAULT 0,
    created_by      BIGINT                   NOT NULL DEFAULT 0,
    updated_by      BIGINT                   NOT NULL DEFAULT 0,
    created_role_id BIGINT                   NOT NULL DEFAULT 0,
    full_name       VARCHAR(100)             NOT NULL,
    employee_no     VARCHAR(50)              NOT NULL,
    username        VARCHAR(100)             NOT NULL,
    password        VARCHAR(255)             NOT NULL,
    gender          SMALLINT                 NOT NULL DEFAULT 0,
    phone_number    VARCHAR(20)              NOT NULL DEFAULT '',
    email           VARCHAR(100)             NOT NULL DEFAULT '',
    enabled         BOOLEAN                  NOT NULL DEFAULT TRUE,
    remark          TEXT                     NOT NULL DEFAULT '',
    position_id     BIGINT                   NOT NULL DEFAULT 0,
    role_id         BIGINT                   NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_employee_no ON users (employee_no) WHERE deleted_at = 0;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users (username) WHERE deleted_at = 0;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_number ON users (phone_number) WHERE deleted_at = 0;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email) WHERE deleted_at = 0;

COMMENT ON TABLE users IS '用户表';
COMMENT ON COLUMN users.id IS '主键';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';
COMMENT ON COLUMN users.deleted_at IS '删除时间';
COMMENT ON COLUMN users.created_by IS '创建人ID';
COMMENT ON COLUMN users.updated_by IS '更新人ID';
COMMENT ON COLUMN users.created_role_id IS '创建时角色ID';
COMMENT ON COLUMN users.full_name IS '用户姓名';
COMMENT ON COLUMN users.employee_no IS '工号';
COMMENT ON COLUMN users.username IS '用户名';
COMMENT ON COLUMN users.password IS '密码';
COMMENT ON COLUMN users.gender IS '性别';
COMMENT ON COLUMN users.phone_number IS '手机号';
COMMENT ON COLUMN users.email IS '邮箱';
COMMENT ON COLUMN users.enabled IS '状态';
COMMENT ON COLUMN users.remark IS '备注';
COMMENT ON COLUMN users.position_id IS '岗位';
COMMENT ON COLUMN users.role_id IS '角色';

INSERT INTO users (full_name, employee_no, username, password, gender, phone_number, email, enabled, position_id,
                   role_id)
VALUES ('超级管理员',
        '0000',
        'admin',
        '$2a$10$VN1FxiFmk1gm4txt2VIJxOLNFKOhyFqm1.EQ2B4rH746u6fQwJOia',
        0,
        '13888888888',
        'admin@example.com',
        TRUE,
        0,
        1)
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
