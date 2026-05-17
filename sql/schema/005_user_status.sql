-- +goose Up
ALTER TABLE users
ADD COLUMN is_red BOOLEAN NOT NULL DEFAULT 'false';

-- +goose Down
ALTER TABLE users
DROP COLUMN is_red;
