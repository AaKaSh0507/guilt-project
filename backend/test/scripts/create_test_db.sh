#!/usr/bin/env bash

set -e

DBNAME="guiltmachine_test_$(date +%s)"
export TEST_DB_NAME="${DBNAME}"
export TEST_DB_URL="postgres://guilt:guiltpass@localhost:5433/${DBNAME}?sslmode=disable"

psql "postgres://guilt:guiltpass@localhost:5433/postgres" -c "CREATE DATABASE ${DBNAME};"

migrate -path ../migrations -database "$TEST_DB_URL" up

echo "$TEST_DB_URL"
