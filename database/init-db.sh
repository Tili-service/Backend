#!/bin/bash

set -e

files=(
    "01-account.sql"
    "02-licence.sql"
    "03-store.sql"
    "04-image.sql"
    "05-payment.sql"
    "06-sales.sql"
    "07-catalog.sql"
    "08-profile.sql"
    "09-categories.sql"
    "10-salehistory.sql"
    "11-item.sql"
)

for f in "${files[@]}"; do
    echo "Importing $f..."
    psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "/docker-entrypoint-initdb.d/tables/$f"
done