# Radar do Povo

**Radar do Povo** é um monorepo open source para distribuição, organização e acessibilidade de dados públicos do governo do Brasil.

O repositório foi estruturado como se todas as APIs/ETLs de ingestão dos datasets públicos já fizessem parte da mesma plataforma, compartilhando a mesma filosofia operacional:

- **Go em produção** para APIs e pipelines de ingestão
- **PostgreSQL** como banco-alvo para quem deseja injetar os dados em banco próprio
- **ETLs idempotentes e retomáveis**
- **Padrão único de observabilidade, importação e versionamento**
- **Monorepo orientado por datasets**, com cada pasta representando uma ingestão/API específica

## Objetivo

Padronizar a ingestão de datasets públicos brasileiros em uma arquitetura consistente, auditável e fácil de evoluir, permitindo que qualquer pessoa ou organização:

- baixe dados oficiais de fontes públicas;
- normalize e valide arquivos CSV/ZIP/JSON/XML;
- injete em PostgreSQL com segurança;
- exponha APIs próprias a partir da base carregada;
- mantenha rastreabilidade, retomada e controle de imports.

## Filosofia do repositório

Cada dataset possui uma estrutura semelhante de projeto, inspirada em pipelines como:

- download oficial do dataset;
- validação do artefato;
- extração e parsing em stream;
- normalização de headers e campos;
- persistência em PostgreSQL;
- controle de progresso em tabela `imports`;
- proteção contra duplicação com chaves sintéticas e `ON CONFLICT DO NOTHING`;
- logs estruturados e métricas de throughput.

Ou seja: o repositório é organizado **como uma coleção de APIs/ETLs de ingestão com arquitetura parecida entre si**.

## Estrutura

```text
radardopovo/
├── .github/
├── apis/
│   ├── emendas-unico/
│   ├── viagens-a-servico/
│   ├── bolsa-familia-pagamentos/
│   ├── auxilio-brasil/
│   ├── cpgf/
│   ├── convenios-federais/
│   ├── servidores-civis/
│   ├── pep/
│   ├── ceis-cnep/
│   └── siafi-despesas/
├── docs/
├── scripts/
└── templates/
```

## O que existe em cada pasta de dataset

Cada pasta em `apis/` representa uma API/ETL de ingestão com a mesma base estrutural:

- `cmd/api/` para entrypoints
- `internal/app/` para orquestração
- `internal/config/` para configuração
- `internal/domain/` para tipos e regras de negócio
- `internal/infra/http/` para clientes HTTP e downloaders
- `internal/infra/postgres/` para persistência e bulk ingest
- `internal/infra/storage/` para arquivos temporários/extração
- `internal/observability/` para logs, métricas e tracing futuros
- `internal/pipeline/` para etapas do ETL
- `tests/` para testes unitários e de integração
- `configs/` para exemplos de configuração

## Catálogo inicial de datasets no monorepo

Ver [docs/CATALOG.md](docs/CATALOG.md).

## Padrões principais

- Banco-alvo: PostgreSQL
- Linguagem base: Go
- Imports com resumibilidade
- Idempotência por chave de linha ou hash sintético
- Dados oficiais sempre identificados pela fonte original
- Sem dependência obrigatória de SaaS proprietário
- Branding e documentação padronizados para open source

## Status do projeto

Este repositório representa a **base organizacional final** do Radar do Povo para múltiplos datasets públicos. As APIs compartilham um modelo estrutural comum e devem evoluir dataset a dataset, mantendo consistência arquitetural.

## Como adicionar um novo dataset

```bash
bash scripts/bootstrap-dataset.sh novo-dataset
```

Isso cria a pasta do dataset com a estrutura padrão do monorepo.

## Licença

MIT. Veja [LICENSE](LICENSE).
