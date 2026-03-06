#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "uso: bash scripts/bootstrap-dataset.sh <dataset-slug>"
  exit 1
fi

DATASET="$1"
TARGET="apis/${DATASET}"

if [ -d "$TARGET" ]; then
  echo "erro: $TARGET já existe"
  exit 1
fi

mkdir -p "$TARGET"/{cmd/api,internal/{app,config,domain,infra/http,infra/postgres,infra/storage,observability,pipeline},configs,migrations,tests}

cat > "$TARGET/README.md" <<EOT
# ${DATASET}

Dataset do monorepo Radar do Povo.

## Fonte oficial

Preencher.

## Objetivo

Preencher.

## Estrutura

- cmd/api
- internal/app
- internal/config
- internal/domain
- internal/infra/http
- internal/infra/postgres
- internal/infra/storage
- internal/observability
- internal/pipeline
- migrations
- tests
EOT

cat > "$TARGET/.env.example" <<EOT
APP_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_NAME=${DATASET//-/_}
DB_USER=postgres
DB_PASS=postgres
DB_SSLMODE=disable
EOT

cat > "$TARGET/go.mod" <<EOT
module github.com/radardopovo/${DATASET}

go 1.24
EOT

echo "dataset criado em $TARGET"
