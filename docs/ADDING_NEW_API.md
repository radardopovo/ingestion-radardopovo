# Como adicionar uma nova API

Este guia existe para quem quiser criar um novo importador no repositorio sem alterar os modulos ja prontos.

## Objetivo

A nova API deve parecer "nativa" dentro do monorepo:

- mesma estrutura de pastas;
- mesma filosofia de idempotencia e resume;
- README local completo;
- isolamento em relacao aos outros datasets.

## Passo 1. Gere o scaffold

```bash
bash scripts/bootstrap-dataset.sh meu-dataset
```

O slug deve ser novo e, para contribuicoes novas, use kebab-case ASCII.

## Passo 2. Defina o contrato do dataset

Antes de escrever ETL, documente:

- fonte oficial;
- URL ou criterio de descoberta;
- granularidade de carga;
- formato dos arquivos;
- tabelas de destino;
- chave de deduplicacao;
- tabela de controle de imports.

## Passo 3. Implemente seguindo o padrao real

Use as APIs existentes como referencia de construcao:

- `bolsa-familia` e `cartão-de-pagamento` para cargas mensais;
- `viagens-a-servico` para carga anual com varios CSVs;
- `emendas-unico` para carga unica com multiplas etapas.

## Passo 4. Preencha o README local

O README do novo modulo deve conter, no minimo:

- descricao do dataset;
- fonte oficial;
- pre-requisitos;
- configuracao;
- build;
- comandos de execucao;
- flags suportadas;
- tabelas criadas;
- consulta de resume/imports;
- notas de performance;
- troubleshooting, se aplicavel.

## Passo 5. Atualize a documentacao compartilhada

Ao adicionar uma nova API:

1. inclua o modulo em [`docs/CATALOG.md`](./CATALOG.md);
2. ajuste o README raiz se a lista de APIs mudou;
3. revise se houve algum novo padrao que mereca entrar em `docs/`.

## Checklist final

- estrutura segue `cmd/importer` + `internal/...`;
- `.env.example` esta completo;
- `Makefile` funciona;
- `go.mod` nomeia o modulo corretamente;
- README local permite uso sem ler o codigo;
- logs e tabela de imports ajudam a operar a carga;
- a API nao depende de outro dataset para funcionar.
