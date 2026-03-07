# Template de API/ETL Go + PostgreSQL

Este template documenta o padrao real esperado para novos modulos do monorepo.

Ele nao representa uma API pronta. Ele serve como referencia estrutural para quem for adicionar uma nova ingestao.

## Estrutura esperada

```text
api-go-etl/
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

## Responsabilidades das pastas

- `cmd/importer`: ponto de entrada do binario
- `internal/config`: flags e envs
- `internal/csvx`: leitura de CSV
- `internal/db`: conexao, migrations, bulk load e imports
- `internal/etl`: orquestracao da carga
- `internal/httpx`: download com timeout e retries
- `internal/logger`: logs estruturados
- `internal/parse`: parsing de valores, datas, hashes e afins
- `internal/zipx`: extracao de ZIP
- `data/`: arquivos baixados e extraidos
- `certs/`: certificados locais

## Como criar um novo modulo

Use o script de bootstrap na raiz do repositorio:

```bash
bash scripts/bootstrap-dataset.sh meu-dataset
```

Depois compare o scaffold gerado com as APIs prontas em `apis/` e adapte a implementacao ao tipo de dataset.
