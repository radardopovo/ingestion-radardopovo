<p align="center">
  <img src="assets/banner.png" alt="Radar do Povo" width="100%">
</p>

# Radar do Povo

Repositorio open source para ingestao de dados publicos do Brasil em bancos externos.

O objetivo deste monorepo e servir como repositorio mae de ETLs em Go, organizados por fonte publica, para que qualquer pessoa ou time possa:

- executar as APIs de ingestao ja prontas;
- carregar os dados em seu proprio PostgreSQL;
- estudar o padrao usado nas APIs existentes;
- criar novas ingestoes seguindo a mesma arquitetura.

## O que existe hoje

Atualmente o repositorio possui 4 APIs de ingestao prontas em [`apis/`](apis/):

| API | Fonte publica | Granularidade | README local |
| --- | --- | --- | --- |
| `bolsa-familia` | Portal da Transparencia | mensal | [apis/bolsa-familia/README.md](apis/bolsa-familia/README.md) |
| `cartão-de-pagamento` | Portal da Transparencia / CPGF | mensal | [apis/cartão-de-pagamento/README.md](apis/cartão-de-pagamento/README.md) |
| `emendas-unico` | Portal da Transparencia / Emendas UNICO | carga unica | [apis/emendas-unico/README.md](apis/emendas-unico/README.md) |
| `viagens-a-servico` | Portal da Transparencia | anual | [apis/viagens-a-servico/README.md](apis/viagens-a-servico/README.md) |

Cada modulo e independente. Isso permite que um usuario use apenas a ingestao que precisa, com banco proprio, sem depender das demais.

## Como o monorepo foi organizado

As APIs atuais seguem um padrao tecnico bem claro:

- `cmd/importer` como ponto de entrada;
- `internal/config` para flags e variaveis de ambiente;
- `internal/httpx` para download e HTTP resiliente;
- `internal/etl` para orquestracao de etapas;
- `internal/db` para conexao, migrations, controle de imports e carga em lote;
- `internal/csvx`, `internal/zipx` e `internal/parse` para leitura, extracao e parsing;
- `internal/logger` para logs estruturados;
- `data/` e `certs/` para artefatos locais;
- README local por API explicando fonte, tabelas, comandos e parametros.

Esse padrao nao foi criado apenas para "organizar pastas". Ele existe para garantir:

- idempotencia;
- retomada apos falha;
- rastreabilidade operacional;
- isolamento entre datasets;
- facilidade para adicionar novos conectores sem acoplamento entre modulos.

## Inicio rapido

1. Escolha uma API em [`apis/`](apis/).
2. Leia o README local do modulo.
3. Copie o `.env.example` da API para `.env`.
4. Configure o PostgreSQL de destino.
5. Execute o importador com `go run ./cmd/importer` ou `make run`, conforme o modulo.

Exemplo com `viagens-a-servico`:

```bash
cd apis/viagens-a-servico
cp .env.example .env
go run ./cmd/importer --from=2011
```

## Criando uma nova API de ingestao

Para criar um novo modulo no mesmo padrao das APIs prontas:

```bash
bash scripts/bootstrap-dataset.sh meu-dataset
```

O script cria a estrutura base do modulo em `apis/meu-dataset/` com:

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

## Estrutura do repositorio

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
- banco controlado pelo proprio usuario;
- sem acoplamento entre datasets;
- schema e regras perto da API correspondente;
- documentacao local obrigatoria;
- padrao repetivel para facilitar contribuicoes open source.

## Documentacao

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

## Contribuicoes

Novas APIs sao bem-vindas, desde que sigam o padrao de ingestao e a documentacao do repositorio.

Antes de abrir PR:

- confira a estrutura do modulo;
- documente a fonte oficial;
- explique a estrategia de resume/idempotencia;
- registre schema, tabelas e comandos de execucao;
- atualize a documentacao do catalogo.

## Licenca

MIT. Veja [LICENSE](LICENSE).
