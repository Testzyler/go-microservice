#!/bin/bash
set -e

# Create additional databases for microservices
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE auth;
    
    -- Grant privileges
    GRANT ALL PRIVILEGES ON DATABASE auth TO $POSTGRES_USER;
EOSQL

echo "Additional databases created: auth"