# viagens-etl

Projeto oficial Radar do Povo: [radardopovo.com](https://radardopovo.com)

ETL em Go para importar **Viagens a Servico** do Portal da Transparencia para AWS RDS PostgreSQL com TLS (`sslmode=verify-full`), idempotencia e retomada por ano.

## 1) Pre-requisitos

- Go 1.23+
- Acesso ao AWS RDS PostgreSQL
- Certificado CA da AWS RDS no servidor

## 2) Obter certificado CA do RDS

```bash
curl -o /certs/global-bundle.pem \
  https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem
```

## 3) Configurar variaveis de ambiente

```bash
cp .env.example .env
# edite .env com suas credenciais
export $(grep -v '^#' .env | xargs)
```

## 4) Rodar importacao completa (2011 -> ano atual)

```bash
go run ./cmd/importer --from=2011
```

## 5) Importar apenas um ano

```bash
go run ./cmd/importer --year=2023
```

## 6) Retomar apos falha (automatico)

```sql
SELECT ano, status, last_step, error_msg
FROM imports
ORDER BY ano;
```

## 7) Forcar reimportacao

```bash
go run ./cmd/importer --year=2022 --force
```

## 8) Schema e relacionamentos

Relacionamento textual:

- `viagens.processo_id` = chave central (pai)
- `trechos.processo_id` referencia o processo de viagem
- `passagens.processo_id` referencia o processo de viagem
- `pagamentos.processo_id` referencia o processo de viagem
- `imports.ano` controla estado da carga por ano (resume/idempotencia)

Tabelas principais:

- `viagens`
- `trechos`
- `passagens`
- `pagamentos`
- `imports`

## 9) Encoding ISO-8859-1

Os CSVs sao lidos em stream e convertidos para UTF-8 com `golang.org/x/text/encoding/charmap`.

Para checar encoding no terminal Linux:

```bash
file -bi data/extracted/2019/*.csv
```

## 10) Parsing de dinheiro (centavos)

Exemplos:

- `1.234,56` -> `123456`
- `1234,56` -> `123456`
- `1234.56` -> `123456`
- `0,00` -> `0`
- `Sem informacao` ou vazio -> `NULL`
- `-500,00` -> `-50000`

## 11) Dicas de performance

- Modo padrao usa `COPY` com merge por `ON CONFLICT`, normalmente mais rapido que `INSERT`.
- Use `--insert-mode` apenas como fallback.
- Ajuste `--batch-size`, `--chunk-size` e `--writers` conforme CPU/IO da VPS e classe da RDS.
- Pool via env:
  - `DB_MAX_OPEN_CONNS`
  - `DB_MAX_IDLE_CONNS`
  - `DB_CONN_MAX_LIFETIME_MIN`
- Em bulk, a sessao usa `SET synchronous_commit = off` para maior throughput.

Exemplo agressivo para RDS pequena:

```bash
go run ./cmd/importer --from=2011 --writers=2 --batch-size=2000 --chunk-size=10000
```

## 12) Troubleshooting TLS

Erros comuns:

- `certificate verify failed`: caminho de `DB_SSL_CA` invalido ou sem permissao de leitura.
- `x509: certificate signed by unknown authority`: CA bundle ausente/desatualizado.
- `hostname mismatch`: `DB_HOST` diferente do endpoint real do RDS.

Checklist rapido:

1. Confirme `DB_SSL_CA=/certs/global-bundle.pem`.
2. Teste conexao com `psql` usando `sslmode=verify-full`.
3. Valide se `DB_HOST` e o endpoint correto da instancia RDS.

## Fonte de dados

- Endpoint anual: `https://portaldatransparencia.gov.br/download-de-dados/viagens/{ANO}`
- O ETL detecta os 4 CSVs por nome:
  - `Viagem`
  - `Trecho`
  - `Passagem`
  - `Pagamento`

## Build e validacao

```bash
make build
make vet
make test
```

## Logging

Campos padrao:

- `source=radardopovo.com`
- `service=viagens-etl`
- `run_id=<unico>`

Para log em JSON:

```bash
go run ./cmd/importer --from=2011 --log-json
```
