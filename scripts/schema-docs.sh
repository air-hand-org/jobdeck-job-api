#!/usr/bin/env bash
set -euo pipefail

# Unique projects isolate concurrent runs; never connect to an existing database.
work=$(mktemp -d)
compose=(docker compose -f scripts/compose.schema-docs.yaml -p "schema-docs-$(basename "$work" | tr '[:upper:].' '[:lower:]-')")
cleanup() {
  "${compose[@]}" down --volumes --remove-orphans >/dev/null || true
  rm -rf "$work"
}
trap cleanup EXIT
"${compose[@]}" up --detach --wait --wait-timeout 90
port=$("${compose[@]}" port postgres 5432)
dsn="postgres://schema_docs:schema_docs@${port}/jobdeck_job?sslmode=disable"
goose -dir db/job/migrations postgres "$dsn" up
tbls doc --config .tbls.yml --dsn "$dsn" --rm-dist
# Server build/architecture metadata is not schema and differs across machines.
jq 'del(.driver.database_version)' docs/schema/schema.json > "$work/schema.json"
mv "$work/schema.json" docs/schema/schema.json
