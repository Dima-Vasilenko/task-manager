#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 \
    --username "$POSTGRES_USER" \
    --dbname "$POSTGRES_DB" \
    --set app_password="$APP_DB_PASSWORD" \
    --set migrator_password="$MIGRATOR_DB_PASSWORD" <<-EOSQL

CREATE ROLE task_manager_migrator
    LOGIN
    PASSWORD :'migrator_password';

CREATE ROLE task_manager_app
    LOGIN
    PASSWORD :'app_password';

GRANT CONNECT ON DATABASE ${POSTGRES_DB}
    TO task_manager_migrator, task_manager_app;

CREATE SCHEMA app AUTHORIZATION task_manager_migrator;

GRANT USAGE ON SCHEMA app TO task_manager_app;

GRANT SELECT, INSERT, UPDATE, DELETE
    ON ALL TABLES IN SCHEMA app
    TO task_manager_app;

GRANT USAGE, SELECT
    ON ALL SEQUENCES IN SCHEMA app
    TO task_manager_app;

ALTER DEFAULT PRIVILEGES FOR ROLE task_manager_migrator IN SCHEMA app
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO task_manager_app;

ALTER DEFAULT PRIVILEGES FOR ROLE task_manager_migrator IN SCHEMA app
    GRANT USAGE, SELECT ON SEQUENCES TO task_manager_app;

EOSQL