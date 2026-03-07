# Contribuindo

Este repositorio foi organizado para funcionar como um monorepo open source de APIs de ingestao de dados publicos do Brasil.

O foco das contribuicoes deve ser manter o padrao atual dos modulos e facilitar a vida de quem vai usar ou estender o projeto.

## O que pode ser contribuido

- novas APIs de ingestao em `apis/<dataset>`;
- melhorias na documentacao;
- correcoes de bugs;
- testes;
- ajustes de observabilidade, importacao e resume;
- padronizacao de scripts e templates.

## Regras basicas

- nao acople datasets entre si sem necessidade real;
- preserve a independencia operacional de cada API;
- documente claramente a fonte oficial do dado;
- mantenha imports idempotentes e retomaveis;
- deixe tabelas, flags e variaveis de ambiente explicitas;
- nao altere semanticamente dado publico sem explicar a normalizacao.

## Padrao esperado para novas APIs

Novos modulos devem seguir a estrutura real ja usada pelas APIs existentes:

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

Use [`docs/ADDING_NEW_API.md`](docs/ADDING_NEW_API.md) como checklist.

## Checklist para Pull Request

Todo PR deve:

1. explicar qual dataset foi afetado;
2. informar a fonte oficial;
3. descrever a granularidade da carga;
4. listar flags, envs e pre-requisitos novos;
5. explicar como a idempotencia e o resume funcionam;
6. atualizar README local e, se necessario, [`docs/CATALOG.md`](docs/CATALOG.md);
7. evitar mudancas desnecessarias em modulos nao relacionados.

## Convencao de commits

Sugestao:

- `feat(dataset): adiciona novo importador`
- `fix(dataset): corrige parser`
- `docs: atualiza guia de contribuicao`
- `chore: ajusta scaffold`

## Validacao local recomendada

Dentro do modulo alterado, rode o maximo possivel:

```bash
go test ./...
go vet ./...
go build ./cmd/importer
```

Se houver `Makefile`, tambem vale usar:

```bash
make test
make vet
make build
```

## Boas praticas para novos datasets

- prefira slugs novos em kebab-case ASCII, sem espacos e sem acentos;
- mantenha o nome tecnico do dataset e o nome oficial da fonte no README;
- isole o schema/tabela de controle de imports do proprio modulo;
- trate CSV, ZIP, encoding e parsing dentro do modulo;
- registre comandos de execucao e exemplos reais.
