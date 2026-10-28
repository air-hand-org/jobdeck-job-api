#!/usr/bin/env bash
set -euo pipefail

# Always create a private, disposable database; never accept a connection URI.
pg_bin=${PG_BIN:-$(pg_config --bindir)}
if [[ $("$pg_bin/postgres" --version) != *" 18."* ]]; then
  echo 'Schema documentation requires PostgreSQL 18.' >&2
  exit 1
fi
work=$(mktemp -d)
cleanup() {
  "$pg_bin/pg_ctl" -D "$work/data" -m immediate -w stop >/dev/null 2>&1 || true
  rm -rf "$work"
}
trap cleanup EXIT
"$pg_bin/initdb" -D "$work/data" --username=schema_docs --auth=trust --locale=C --encoding=UTF8 >/dev/null
"$pg_bin/pg_ctl" -D "$work/data" -l "$work/postgres.log" \
  -o "-k $work -c listen_addresses=''" -w start >/dev/null
"$pg_bin/createdb" -h "$work" -U schema_docs jobdeck_job
# goose accepts libpq URLs; tbls uses dburl's Unix-socket URL syntax.
goose -dir db/job/migrations postgres "postgres://schema_docs@localhost/jobdeck_job?host=$work&sslmode=disable" up
tbls doc --config .tbls.yml --dsn "postgres://schema_docs@$work/jobdeck_job?sslmode=disable" --rm-dist
# Server build/architecture metadata is not schema and differs across machines.
jq 'del(.driver.database_version)' docs/schema/schema.json > "$work/schema.json"
mv "$work/schema.json" docs/schema/schema.json
