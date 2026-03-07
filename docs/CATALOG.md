# Catalogo de APIs

Este catalogo lista os modulos de ingestao atualmente disponiveis no monorepo.

## APIs disponiveis

### `bolsa-familia`

- fonte: Portal da Transparencia / Novo Bolsa Familia
- granularidade: mensal (`YYYYMM`)
- tabelas principais: `bolsa_familia`, `imports_bolsa_familia`
- README: [../apis/bolsa-familia/README.md](../apis/bolsa-familia/README.md)

### `cartão-de-pagamento`

- fonte: Portal da Transparencia / Cartao de Pagamento do Governo Federal (CPGF)
- granularidade: mensal (`YYYYMM`)
- tabelas principais: `cpgf`, `imports_cpgf`
- README: [../apis/cartão-de-pagamento/README.md](../apis/cartão-de-pagamento/README.md)

### `emendas-unico`

- fonte: Portal da Transparencia / Emendas Parlamentares (UNICO)
- granularidade: carga unica com 3 CSVs
- tabelas principais: `emendas`, `emendas_por_favorecido`, `emendas_convenios`, `imports`
- README: [../apis/emendas-unico/README.md](../apis/emendas-unico/README.md)

### `viagens-a-servico`

- fonte: Portal da Transparencia / Viagens a Servico
- granularidade: anual
- tabelas principais: `viagens`, `trechos`, `passagens`, `pagamentos`, `imports`
- README: [../apis/viagens-a-servico/README.md](../apis/viagens-a-servico/README.md)

## O que todo modulo deve documentar

Cada API nova deve registrar no README local:

- nome do dataset;
- fonte oficial;
- formato de entrada;
- granularidade da carga;
- tabelas de destino;
- comandos de execucao;
- flags suportadas;
- variaveis de ambiente;
- estrategia de idempotencia e resume;
- observacoes de performance e troubleshooting, quando fizer sentido.
