CREATE TABLE IF NOT EXISTS users
(
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100)             NOT NULL,
    email      VARCHAR(100)             NOT NULL UNIQUE,
    password   VARCHAR(100)             NOT NULL,
    enabled    BOOLEAN                  NOT NULL DEFAULT TRUE,
    role       VARCHAR(20)              NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
