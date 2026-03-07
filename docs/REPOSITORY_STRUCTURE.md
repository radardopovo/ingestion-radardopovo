# Estrutura do repositorio

O monorepo e orientado por dataset. Cada pasta em `apis/` representa uma API de ingestao independente.

## Visao geral

```text
.
|-- apis/
|   |-- bolsa-familia/
|   |-- cartão-de-pagamento/
|   |-- emendas-unico/
|   `-- viagens-a-servico/
|-- docs/
|   |-- architecture/
|   `-- standards/
|-- scripts/
|-- templates/
|-- README.md
`-- CONTRIBUTING.md
```

## O que fica em cada area

### `apis/`

Contem os importadores prontos e futuros modulos de ingestao.

Cada modulo deve ser autocontido:

- propria configuracao;
- propria documentacao;
- proprio fluxo ETL;
- proprio schema/tabelas;
- proprio controle de progresso de importacao.

### `docs/`

Documentacao compartilhada do monorepo:

- catalogo atual;
- padrao de ingestao;
- padrao de uso com PostgreSQL;
- convencoes de naming, dados e observabilidade;
- guias para uso e extensao.

### `scripts/`

Automacoes auxiliares do repositorio, incluindo o bootstrap para novos datasets.

### `templates/`

Material de referencia para scaffolding e padrao estrutural.

## Estrutura interna esperada por API

```text
apis/<dataset>/
|-- cmd/importer/
|-- internal/config/
|-- internal/csvx/
|-- internal/db/
|-- internal/etl/
|-- internal/httpx/
|-- internal/logger/
|-- internal/parse/
|-- internal/zipx/
|-- certs/
|-- data/
|-- .env.example
|-- .gitignore
|-- CONTRIBUTING.md
|-- LICENSE
|-- Makefile
|-- README.md
`-- go.mod
```

Essa estrutura foi derivada do padrao real das APIs existentes. Novos modulos devem seguir esse desenho, salvo justificativa tecnica forte.
