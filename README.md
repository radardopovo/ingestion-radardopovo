<p align="center">
  <img src="assets/banner.png" alt="Radar do Povo" width="100%">
</p>

<p align="center">
  Repositório open source para ingestão de dados públicos do Brasil em bancos externos.
</p>

<p align="center">
  <a href="mailto:radardopovo@proton.me">radardopovo@proton.me</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="MIT License">
  <img src="https://img.shields.io/badge/go-ingestion-00ADD8" alt="Go">
  <img src="https://img.shields.io/badge/postgresql-supported-336791" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/open%20source-yes-brightgreen" alt="Open Source">
  <img src="https://img.shields.io/badge/status-active-success" alt="Status">
</p>

---

## Visão geral

O **Radar do Povo** é um monorepo open source voltado à ingestão de dados públicos do Brasil em bancos externos, com foco em **ETLs em Go**, organização por dataset e execução independente por módulo.

A proposta do repositório é servir como **repositório-mãe de APIs de ingestão**, permitindo que qualquer pessoa, time ou organização possa:

- executar ingestões já prontas;
- carregar os dados em PostgreSQL próprio;
- estudar o padrão técnico adotado nas APIs existentes;
- criar novos módulos seguindo a mesma arquitetura;
- reaproveitar um modelo consistente de download, parsing, persistência e observabilidade.

## O que existe hoje

Atualmente o repositório possui 4 APIs de ingestão prontas em [`apis/`](apis/):

| API | Fonte pública | Granularidade | README local |
| --- | --- | --- | --- |
| `bolsa-familia` | Portal da Transparência | mensal | [apis/bolsa-familia/README.md](apis/bolsa-familia/README.md) |
| `cartão-de-pagamento` | Portal da Transparência / CPGF | mensal | [apis/cartão-de-pagamento/README.md](apis/cartão-de-pagamento/README.md) |
| `emendas-unico` | Portal da Transparência / Emendas UNICO | carga única | [apis/emendas-unico/README.md](apis/emendas-unico/README.md) |
| `viagens-a-servico` | Portal da Transparência | anual | [apis/viagens-a-servico/README.md](apis/viagens-a-servico/README.md) |

Cada módulo é independente. Isso permite que o usuário utilize apenas a ingestão que precisa, com banco próprio, sem depender das demais APIs do monorepo.

## Por que este projeto existe

Bases públicas oficiais costumam vir com formatos diferentes, granularidades diferentes, problemas de padronização e fluxos de ingestão pouco reaproveitáveis.

O Radar do Povo existe para oferecer uma base open source clara e repetível para:

- baixar dados de fontes públicas oficiais;
- validar artefatos como CSV, ZIP e arquivos relacionados;
- normalizar campos, headers e estruturas de importação;
- persistir dados em PostgreSQL de forma controlada;
- manter rastreabilidade operacional e retomada após falha;
- facilitar a criação de novas ingestões no mesmo padrão.

## Como o monorepo foi organizado

As APIs atuais seguem um padrão técnico comum:

- `cmd/importer` como ponto de entrada;
- `internal/config` para flags e variáveis de ambiente;
- `internal/httpx` para download e HTTP resiliente;
- `internal/etl` para orquestração de etapas;
- `internal/db` para conexão, migrations, controle de imports e carga em lote;
- `internal/csvx`, `internal/zipx` e `internal/parse` para leitura, extração e parsing;
- `internal/logger` para logs estruturados;
- `data/` e `certs/` para artefatos locais;
- README local por API explicando fonte, tabelas, comandos e parâmetros.

Esse padrão não existe apenas para organizar pastas. Ele foi pensado para garantir:

- idempotência;
- retomada após falha;
- rastreabilidade operacional;
- isolamento entre datasets;
- facilidade para adicionar novos conectores sem acoplamento entre módulos.

## Início rápido

1. Escolha uma API em [`apis/`](apis/).
2. Leia o README local do módulo.
3. Copie o `.env.example` da API para `.env`.
4. Configure o PostgreSQL de destino.
5. Execute o importador com `go run ./cmd/importer` ou `make run`, conforme o módulo.

Exemplo com `viagens-a-servico`:

```bash
cd apis/viagens-a-servico
cp .env.example .env
go run ./cmd/importer --from=2011
```

## Criando uma nova API de ingestão

Para criar um novo módulo no mesmo padrão das APIs prontas:

```bash
bash scripts/bootstrap-dataset.sh meu-dataset
```

O script cria a estrutura base do módulo em `apis/meu-dataset/` com:

- `cmd/importer`
- `internal/config`
- `internal/csvx`
- `internal/db`
- `internal/etl`
- `internal/httpx`
- `internal/logger`
- `internal/parse`
- `internal/zipx`
- `data/`
- `certs/`
- arquivos iniciais de README, Makefile, `.env.example`, `go.mod` e `LICENSE` quando existir na raiz

Antes de implementar um novo dataset, leia:

- [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)
- [docs/ADDING_NEW_API.md](docs/ADDING_NEW_API.md)
- [docs/CATALOG.md](docs/CATALOG.md)
- [docs/architecture/INGESTION_STANDARD.md](docs/architecture/INGESTION_STANDARD.md)
- [docs/architecture/POSTGRES_PATTERN.md](docs/architecture/POSTGRES_PATTERN.md)

## Estrutura do repositório

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
`-- templates/
```

## Filosofia do projeto

- fonte oficial primeiro;
- banco controlado pelo próprio usuário;
- sem acoplamento entre datasets;
- schema e regras perto da API correspondente;
- documentação local obrigatória;
- padrão repetível para facilitar contribuições open source.

## Documentação

- [docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)
- [docs/CATALOG.md](docs/CATALOG.md)
- [docs/REPOSITORY_STRUCTURE.md](docs/REPOSITORY_STRUCTURE.md)
- [docs/ADDING_NEW_API.md](docs/ADDING_NEW_API.md)
- [docs/architecture/INGESTION_STANDARD.md](docs/architecture/INGESTION_STANDARD.md)
- [docs/architecture/POSTGRES_PATTERN.md](docs/architecture/POSTGRES_PATTERN.md)
- [docs/standards/DATA_POLICY.md](docs/standards/DATA_POLICY.md)
- [docs/standards/NAMING.md](docs/standards/NAMING.md)
- [docs/standards/OBSERVABILITY.md](docs/standards/OBSERVABILITY.md)
- [CONTRIBUTING.md](CONTRIBUTING.md)
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- [SECURITY.md](SECURITY.md)
- [SUPPORT.md](SUPPORT.md)

## Contribuições

Novas APIs são bem-vindas, desde que sigam o padrão de ingestão e a documentação do repositório.

Antes de abrir um PR:

- confira a estrutura do módulo;
- documente a fonte oficial;
- explique a estratégia de resume/idempotência;
- registre schema, tabelas e comandos de execução;
- atualize a documentação do catálogo.

Leia também:

- [CONTRIBUTING.md](CONTRIBUTING.md)
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- [SECURITY.md](SECURITY.md)

## Contato

Para dúvidas gerais, sugestões, colaboração ou contato institucional:

**radardopovo@proton.me**

## Licença

MIT. Veja [LICENSE](LICENSE).
