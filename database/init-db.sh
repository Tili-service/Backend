#!/bin/bash

set -e

files=(
    "account.sql"
    "categories.sql"
    "image.sql"
    "catalogue.sql"
    "payment.sql"
    "sales.sql"
    "store.sql"
    "user.sql"
)

for f in "${files[@]}"; do
    psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "/docker-entrypoint-initdb.d/tables/$f"
done
