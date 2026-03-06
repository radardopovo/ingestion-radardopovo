# Padrão PostgreSQL

O banco-alvo padrão do monorepo é PostgreSQL.

## Recomendações gerais

- usar TLS quando remoto;
- manter migrations versionadas por dataset;
- proteger duplicação com índices e conflitos explícitos;
- usar tabela `imports` para controle operacional;
- evitar dependência de features não portáveis quando não necessário.
