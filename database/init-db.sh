#!/bin/bash

set -e

files=(
    "account.sql"
    "licence.sql"
    "image.sql"
    "store.sql"
    "catalog.sql"
    "categories.sql"
    "item.sql"
    "payment.sql"
    "store.sql"
    "profile.sql"
    "sales.sql"
    "salehistory.sql"
)

for f in "${files[@]}"; do
    psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "/docker-entrypoint-initdb.d/tables/$f"
done
