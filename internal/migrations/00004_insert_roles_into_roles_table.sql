-- +goose Up
-- +goose StatementBegin
INSERT INTO roles (name) 
VALUES ('admin'), ('user')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM roles;
-- +goose StatementEnd
