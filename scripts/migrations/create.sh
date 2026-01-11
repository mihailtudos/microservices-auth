#!/bin/sh
set -e

echo "Enter a migration name"

read migration_name

goose -s create "$migration_name" sql --dir "$GOOSE_MIGRATION_DIR"