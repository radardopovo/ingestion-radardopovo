# Padrao PostgreSQL

O banco de destino padrao do monorepo e PostgreSQL controlado pelo proprio usuario.

## Objetivo

As APIs de ingestao existem para carregar dados publicos em um banco externo do usuario, e nao para forcar um storage central unico.

## Regras gerais

- cada modulo deve definir claramente suas tabelas de destino;
- a estrategia de deduplicacao precisa ser explicita;
- imports devem ser idempotentes;
- reimportacoes devem ser controladas por flags e/ou estado persistido;
- a documentacao deve explicar como acompanhar o progresso.

## Tabela de controle de importacao

As APIs atuais usam duas variacoes validas:

- tabela generica `imports`;
- tabela especifica do dataset, como `imports_cpgf` ou `imports_bolsa_familia`.

Qualquer uma das abordagens e aceitavel, desde que:

- a escolha seja consistente dentro do modulo;
- o README explique a chave de controle;
- o usuario consiga entender status, etapa atual, erro e contagem de linhas.

## Carga em lote

As APIs atuais privilegiam bulk load e insercao eficiente:

- `COPY` quando possivel;
- fallback para `INSERT` quando necessario;
- batches configuraveis;
- workers configuraveis;
- `ON CONFLICT` ou estrategia equivalente para evitar duplicacao.

## Configuracao de conexao

Os modulos atuais usam variaveis como:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASS`
- `DB_NAME`
- `DB_SSL_CA`
- `DB_MAX_OPEN_CONNS`
- `DB_MAX_IDLE_CONNS`
- `DB_CONN_MAX_LIFETIME_MIN`

Quando o banco for remoto, TLS deve ser documentado com clareza.

## O que um novo modulo deve documentar

- schema e tabelas;
- chave primaria ou hash sintetico;
- estrategia de upsert/deduplicacao;
- consulta basica para acompanhar imports;
- impacto esperado de `--force`;
- recomendacoes de performance para bancos pequenos.
