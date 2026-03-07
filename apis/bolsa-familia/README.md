# bolsa-familia-etl

Projeto oficial Radar do Povo: [radardopovo.com](https://radardopovo.com)

ETL em Go para importar **Novo Bolsa Familia** do Portal da Transparencia para PostgreSQL (AWS RDS), com TLS, idempotencia e retomada mensal.

## Fonte de dados

- Endpoint mensal: `https://portaldatransparencia.gov.br/download-de-dados/novo-bolsa-familia/YYYYMM`
- Exemplo: `https://portaldatransparencia.gov.br/download-de-dados/novo-bolsa-familia/202601`
- ZIP esperado com CSV como `202601_NovoBolsaFamilia.csv`

## Colunas do CSV

- `MES COMPETENCIA`
- `MES REFERENCIA`
- `UF`
- `CODIGO MUNICIPIO SIAFI`
- `NOME MUNICIPIO`
- `CPF FAVORECIDO`
- `NIS FAVORECIDO`
- `NOME FAVORECIDO`
- `VALOR PARCELA`

## Escopo de importacao

- Faixa padrao: 2016 ate o mes atual
- Dataset mensal (`YYYYMM`)

## Pre-requisitos

- Go 1.23+
- PostgreSQL 17+ (RDS recomendado)
- Certificado CA da AWS RDS

## Certificado CA

```bash
curl -o /certs/global-bundle.pem \
  https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem
```

## Configuracao

```bash
cp .env.example .env
set -a; source .env; set +a
```

## Build

```bash
go mod tidy
go build -o bin/importer ./cmd/importer
go vet ./...
```

## Execucao

Importar tudo:

```bash
./bin/importer --from=2016
```

Importar um ano inteiro:

```bash
./bin/importer --year=2026
```

Importar periodo unico:

```bash
./bin/importer --period=202601
```

Forcar reimportacao:

```bash
./bin/importer --period=202601 --force
```

## Tabelas

- `bolsa_familia`
- `imports_bolsa_familia`

## Resume

```sql
SELECT periodo_yyyymm, status, last_step, rows_bolsa_familia, error_msg
FROM imports_bolsa_familia
ORDER BY periodo_yyyymm;
```

## Performance

Defaults para RDS pequena:

- `DB_MAX_OPEN_CONNS=2`
- `DB_MAX_IDLE_CONNS=2`
- `WRITER_WORKERS=2`
- `BATCH_SIZE=2000`
- `CHUNK_SIZE=10000`

## Logging

Campos padrao:

- `source=radardopovo.com`
- `service=bolsa-familia-etl`
- `run_id=<unico>`
