# Observabilidade

Cada API deve produzir sinais operacionais suficientes para diagnostico e operacao segura.

## Minimo esperado

- logs estruturados;
- identificacao por `run_id`;
- nome do dataset ou servico;
- etapa atual da ingestao;
- contagem de linhas lidas, inseridas e ignoradas;
- duracao total e, quando possivel, throughput.

## Eventos importantes

O usuario deve conseguir identificar pelos logs:

- inicio da execucao;
- inicio e fim do download;
- extracao de arquivos;
- inicio e fim de cada etapa de importacao;
- motivo de erro;
- conclusao ou skip por cache/import ja concluido.

## Estado persistido

Sempre que houver tabela de imports, ela deve permitir responder:

- qual periodo ou ano falhou;
- em qual etapa parou;
- se a carga terminou;
- quantas linhas foram inseridas;
- qual foi o erro registrado.
