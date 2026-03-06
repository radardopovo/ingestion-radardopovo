# Catálogo de datasets

Este catálogo trata o monorepo como uma coleção de APIs/ETLs de ingestão com estrutura semelhante.

## Datasets já previstos na base do repositório

- `emendas-unico`
- `viagens-a-servico`
- `bolsa-familia-pagamentos`
- `auxilio-brasil`
- `cpgf`
- `convenios-federais`
- `servidores-civis`
- `pep`
- `ceis-cnep`
- `siafi-despesas`

## Estrutura esperada por dataset

Cada dataset deve conter:

- descrição do domínio do dado;
- fonte oficial;
- formato de entrada;
- estratégia de parsing;
- esquema PostgreSQL próprio;
- tabela `imports` para progresso;
- README local;
- `.env.example` local quando necessário.
