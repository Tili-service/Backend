#!/bin/bash

set -e

for f in /docker-entrypoint-initdb.d/tables/*.sql; do
    psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "$f"
done
