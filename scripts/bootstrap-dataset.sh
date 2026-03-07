#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "uso: bash scripts/bootstrap-dataset.sh <dataset-slug>"
  exit 1
fi

DATASET="$1"

if ! [[ "$DATASET" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]]; then
  echo "erro: use um slug em kebab-case ASCII, por exemplo: meu-dataset"
  exit 1
fi

TARGET="apis/${DATASET}"
MODULE="github.com/radardopovo/${DATASET}-etl"
DATASET_SNAKE="${DATASET//-/_}"

if [ -d "$TARGET" ]; then
  echo "erro: $TARGET ja existe"
  exit 1
fi

mkdir -p \
  "$TARGET/cmd/importer" \
  "$TARGET/internal/config" \
  "$TARGET/internal/csvx" \
  "$TARGET/internal/db" \
  "$TARGET/internal/etl" \
  "$TARGET/internal/httpx" \
  "$TARGET/internal/logger" \
  "$TARGET/internal/parse" \
  "$TARGET/internal/zipx" \
  "$TARGET/certs" \
  "$TARGET/data"

touch "$TARGET/certs/.gitkeep" "$TARGET/data/.gitkeep"

if [ -f LICENSE ]; then
  cp LICENSE "$TARGET/LICENSE"
fi

cat > "$TARGET/README.md" <<EOF
# ${DATASET}-etl

ETL em Go para importar um dataset publico brasileiro para PostgreSQL.

## Fonte de dados

- fonte oficial: preencher
- URL base: preencher
- granularidade: preencher

## Objetivo

Explique aqui o que esta API baixa, transforma e grava no banco.

## Pre-requisitos

- Go 1.23+
- PostgreSQL
- certificado TLS, se o banco remoto exigir

## Configuracao

\`\`\`bash
cp .env.example .env
\`\`\`

Preencha as variaveis de banco, dados e performance.

## Build

\`\`\`bash
make build
make vet
make test
\`\`\`

## Execucao

\`\`\`bash
go run ./cmd/importer
\`\`\`

Adicione aqui exemplos reais de flags, como \`--from\`, \`--year\`, \`--period\` ou \`--force\`.

## Tabelas

- ${DATASET_SNAKE}
- imports_${DATASET_SNAKE}

## Resume

Documente a consulta SQL usada para acompanhar o progresso da carga.

## Estrutura interna

- \`cmd/importer\`
- \`internal/config\`
- \`internal/csvx\`
- \`internal/db\`
- \`internal/etl\`
- \`internal/httpx\`
- \`internal/logger\`
- \`internal/parse\`
- \`internal/zipx\`
- \`data/\`
- \`certs/\`
EOF

cat > "$TARGET/CONTRIBUTING.md" <<EOF
# Contributing

## Requisitos

- Go 1.23+
- PostgreSQL configurado para o ambiente de desenvolvimento

## Fluxo recomendado

1. implemente a fonte oficial no modulo;
2. preserve idempotencia e resume;
3. atualize o README local;
4. rode \`go test ./...\`, \`go vet ./...\` e \`go build ./cmd/importer\`;
5. documente tabelas, envs e flags.
EOF

cat > "$TARGET/.env.example" <<EOF
# ${DATASET}-etl

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=${DATASET_SNAKE}
DB_SSL_CA=/certs/global-bundle.pem

DB_MAX_OPEN_CONNS=30
DB_MAX_IDLE_CONNS=30
DB_CONN_MAX_LIFETIME_MIN=10

DATA_DIR=./data
BATCH_SIZE=1000
CHUNK_SIZE=5000
WRITER_WORKERS=2
HTTP_TIMEOUT_SEC=120
HTTP_MAX_RETRIES=5
LOG_JSON=false
EOF

cat > "$TARGET/.gitignore" <<'EOF'
.env
bin/
data/extracted/
data/zips/
EOF

cat > "$TARGET/Makefile" <<'EOF'
.PHONY: build vet test run clean

build:
	go build -o bin/importer ./cmd/importer

vet:
	go vet ./...

test:
	go test ./...

run:
	go run ./cmd/importer

clean:
	rm -f bin/importer
EOF

cat > "$TARGET/go.mod" <<EOF
module ${MODULE}

go 1.23
EOF

cat > "$TARGET/cmd/importer/main.go" <<EOF
package main

import "fmt"

func main() {
	fmt.Println("${DATASET}-etl scaffold criado. Implemente config, db, httpx e etl antes de usar em producao.")
}
EOF

cat > "$TARGET/internal/config/config.go" <<'EOF'
package config

// TODO: implementar leitura de flags e variaveis de ambiente.
EOF

cat > "$TARGET/internal/csvx/reader.go" <<'EOF'
package csvx

// TODO: implementar leitura de CSV em stream.
EOF

cat > "$TARGET/internal/db/connect.go" <<'EOF'
package db

// TODO: implementar conexao, migrations e controle de imports.
EOF

cat > "$TARGET/internal/etl/orchestrator.go" <<'EOF'
package etl

// TODO: implementar orquestracao das etapas de download, extracao e importacao.
EOF

cat > "$TARGET/internal/httpx/client.go" <<'EOF'
package httpx

// TODO: implementar cliente HTTP com timeout e retries.
EOF

cat > "$TARGET/internal/logger/logger.go" <<'EOF'
package logger

// TODO: implementar logger estruturado com run_id.
EOF

cat > "$TARGET/internal/parse/parse.go" <<'EOF'
package parse

// TODO: implementar parsing de datas, valores e normalizacoes do dataset.
EOF

cat > "$TARGET/internal/zipx/extract.go" <<'EOF'
package zipx

// TODO: implementar extracao de ZIP quando aplicavel.
EOF

echo "dataset criado em $TARGET"
