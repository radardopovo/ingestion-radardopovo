# cpgf-etl

Projeto oficial Radar do Povo: [radardopovo.com](https://radardopovo.com)

ETL em Go para importar **Cartao de Pagamento do Governo Federal (CPGF)** do Portal da Transparencia para PostgreSQL (AWS RDS), com TLS, idempotencia e retomada.

## Fonte de dados

- Endpoint mensal: `https://portaldatransparencia.gov.br/download-de-dados/cpgf/YYYYMM`
- Exemplo: `https://portaldatransparencia.gov.br/download-de-dados/cpgf/202601`
- Retorno esperado: ZIP (ex.: `202512_CPGF.zip`) contendo CSV (ex.: `202512_CPGF.csv`)

## Escopo de importacao

- Faixa padrao: 2016 ate o mes atual
- Dataset mensal (`YYYYMM`)
- Um unico CSV por periodo

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
# edite credenciais
set -a; source .env; set +a
```

## Build

```bash
go mod tidy
go build -o bin/importer ./cmd/importer
go vet ./...
```

## Execucao

Importar tudo (2016 ate hoje):

```bash
./bin/importer --from=2016
```

Importar um ano inteiro:

```bash
./bin/importer --year=2025
```

Importar um periodo unico:

```bash
./bin/importer --period=202601
```

Forcar reimportacao de um periodo:

```bash
./bin/importer --period=202512 --force
```

Apenas download/extracao:

```bash
./bin/importer --period=202601 --only-download
```

Apenas import (sem baixar):

```bash
./bin/importer --period=202601 --only-import
```

## Tabelas

- `cpgf`
- `imports_cpgf`

### cpgf

Campos principais:

- orgaos (`codigo_orgao_superior`, `codigo_orgao`, `codigo_unidade_gestora`)
- portador (`cpf_portador`, `nome_portador`)
- favorecido (`documento_favorecido`, `nome_favorecido`)
- transacao (`transacao`, `data_transacao`, `valor_transacao_cents`)
- `periodo_yyyymm`
- `id` (SHA-1 sintetico, PK)

### imports_cpgf

Controle de resume por `periodo_yyyymm`:

- `status`: `downloading | extracted | importing | done | error`
- `last_step`: `download | extract | cpgf`
- `rows_cpgf`, `error_msg`, `started_at`, `finished_at`

## Idempotencia

- Insercao em `cpgf` usa `ON CONFLICT (id) DO NOTHING`
- Reexecucao nao duplica registros
- `--force` limpa periodo antes de reimportar

## Resume

Consulta de status:

```sql
SELECT periodo_yyyymm, status, last_step, rows_cpgf, error_msg
FROM imports_cpgf
ORDER BY periodo_yyyymm;
```

## Performance

- Padrao usa `COPY` (mais rapido)
- Fallback: `--insert-mode`
- Defaults para RDS pequena:
  - `DB_MAX_OPEN_CONNS=2`
  - `DB_MAX_IDLE_CONNS=2`
  - `WRITER_WORKERS=2`
  - `BATCH_SIZE=2000`
  - `CHUNK_SIZE=10000`
- Sessao bulk:
  - `synchronous_commit=off`
  - `idle_in_transaction_session_timeout=60s`

## Logging

Campos padrao:

- `source=radardopovo.com`
- `service=cpgf-etl`
- `run_id=<unico>`

Ativar JSON:

```bash
./bin/importer --from=2016 --log-json
```
