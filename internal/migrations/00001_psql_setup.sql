-- +goose Up
-- +goose StatementBegin
SET timezone = 'Europe/London';
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
RESET timezone;
-- +goose StatementEnd