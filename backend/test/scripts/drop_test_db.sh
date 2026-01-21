#!/usr/bin/env bash

set -e

DBNAME="${TEST_DB_NAME}"
if [ -z "$DBNAME" ]; then
  echo "TEST_DB_NAME not set, skip drop"
  exit 0
fi

psql "postgres://guilt:guiltpass@localhost:5433/postgres" -c "DROP DATABASE IF EXISTS ${DBNAME};"
