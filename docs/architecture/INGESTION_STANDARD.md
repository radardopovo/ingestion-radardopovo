# Padrao de ingestao

As APIs deste monorepo seguem um fluxo operacional parecido, mesmo quando o dataset muda.

## Fluxo esperado

1. definir a fonte oficial e a granularidade da carga;
2. baixar o artefato remoto;
3. validar existencia, formato e nome esperado;
4. extrair ZIP quando aplicavel;
5. ler os arquivos em stream;
6. normalizar encoding, cabecalhos e formatos;
7. transformar em tipos adequados para persistencia;
8. gravar em PostgreSQL com estrategia idempotente;
9. registrar progresso de importacao;
10. permitir retomada ou reimportacao controlada.

## Pastas envolvidas no fluxo

- `cmd/importer`: inicializa config, logger, cliente HTTP, banco e orquestrador;
- `internal/config`: flags e envs;
- `internal/httpx`: download e retries;
- `internal/zipx`: extracao;
- `internal/csvx`: leitura de CSV;
- `internal/parse`: datas, dinheiro, hashing e utilitarios de parsing;
- `internal/db`: conexao, migrations, bulk load e estado de importacao;
- `internal/etl`: coordenacao das etapas por dataset;
- `internal/logger`: logs estruturados.

## Principios obrigatorios

- fonte oficial primeiro;
- idempotencia;
- resume apos falha;
- rastreabilidade;
- isolamento por dataset;
- documentacao local suficiente para uso sem adivinhacao.

## O que um novo modulo deve decidir logo no inicio

- qual e a granularidade de importacao: anual, mensal, diario ou carga unica;
- se ha um ou varios arquivos por execucao;
- qual sera a tabela de controle operacional;
- qual sera a chave de deduplicacao;
- quando `--force` deve purgar e reimportar;
- como o modulo identifica cache local reutilizavel.

## Flags comuns encontradas nas APIs atuais

Nem todas as APIs possuem exatamente as mesmas flags, mas este e o conjunto mais recorrente:

- `--from`
- `--to`
- `--year`
- `--period`
- `--force`
- `--only-download`
- `--only-import`
- `--insert-mode`
- `--dry-run`
- `--log-json`

Se um novo modulo nao suportar alguma flag comum, o README local deve explicar o motivo.
