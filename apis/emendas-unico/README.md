# emendas-etl

Projeto oficial Radar do Povo: [radardopovo.com](https://radardopovo.com)

ETL em Go para importar **Emendas Parlamentares (UNICO)** do Portal da Transparencia para AWS RDS PostgreSQL com TLS, idempotencia e retomada.

## Fonte

- URL unica: `https://portaldatransparencia.gov.br/download-de-dados/emendas-parlamentares/UNICO`
- ZIP com 3 CSVs:
  - `EmendasParlamentares.csv`
  - `EmendasParlamentares_PorFavorecido.csv`
  - `EmendasParlamentares_Convenios.csv`

## Pre-requisitos

- Go 1.23+
- PostgreSQL (AWS RDS recomendado)
- Certificado CA da AWS RDS

## Certificado CA

```bash
curl -o /certs/global-bundle.pem \
  https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem
```

## Configuracao

```bash
cp .env.example .env
export $(grep -v '^#' .env | xargs)
```

## Execucao

Importacao completa:

```bash
go run ./cmd/importer
```

Forcar reimportacao:

```bash
go run ./cmd/importer --force
```

Apenas download + extracao:

```bash
go run ./cmd/importer --only-download
```

Apenas import (sem baixar):

```bash
go run ./cmd/importer --only-import
```

Modo agressivo:

```bash
go run ./cmd/importer --writers=2 --batch-size=2000 --chunk-size=10000
```

## Resume

Controle em `imports` por `dataset_key='emendas_unico'`.

```sql
SELECT dataset_key, status, last_step, error_msg, rows_emendas, rows_favorecido, rows_convenio
FROM imports;
```

## Tabelas

- `emendas`
- `emendas_por_favorecido`
- `emendas_convenios`
- `imports`

## Logging

Campos padrao:

- `source=radardopovo.com`
- `service=emendas-etl`
- `run_id=<unico>`

JSON:

```bash
go run ./cmd/importer --log-json
```

## Desenvolvimento

```bash
make build
make test
make vet
```
