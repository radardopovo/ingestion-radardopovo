# Padrão de ingestão

As APIs/ETLs do monorepo seguem, em geral, este fluxo:

1. descoberta/definição da URL oficial;
2. download do artefato;
3. validação básica do arquivo;
4. extração quando aplicável;
5. leitura em stream;
6. normalização de campos e cabeçalhos;
7. transformação tipada;
8. persistência em PostgreSQL;
9. marcação de progresso em `imports`;
10. conclusão idempotente.

## Princípios

- resumibilidade;
- idempotência;
- rastreabilidade;
- logs estruturados;
- isolamento por dataset.
