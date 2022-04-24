CREATE TYPE role_enum AS ENUM ('user', 'admin');
ALTER TABLE users ADD COLUMN role role_enum NOT NULL DEFAULT 'user';