# Como usar as APIs existentes

Este guia resume como executar os modulos ja prontos do repositorio.

## 1. Escolha a API

As APIs atualmente disponiveis estao em:

- `apis/bolsa-familia`
- `apis/cartão-de-pagamento`
- `apis/emendas-unico`
- `apis/viagens-a-servico`

Cada uma possui README proprio com detalhes da fonte, flags e tabelas.

## 2. Entre no modulo

```bash
cd apis/viagens-a-servico
```

## 3. Configure o ambiente

```bash
cp .env.example .env
```

Depois ajuste o arquivo `.env` com:

- host, porta e credenciais do PostgreSQL;
- caminho do certificado TLS, se necessario;
- parametros de pool e performance;
- diretorio local de dados.

## 4. Baixe o certificado do RDS quando aplicavel

Alguns modulos assumem uso com AWS RDS e TLS:

```bash
curl -o /certs/global-bundle.pem \
  https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem
```

Se voce usar outro ambiente, adapte o certificado e os parametros de conexao conforme seu banco.

## 5. Rode a API

Exemplos comuns:

```bash
go run ./cmd/importer --from=2011
go run ./cmd/importer --year=2025
go run ./cmd/importer --period=202601
go run ./cmd/importer --period=202601 --force
```

## 6. Valide o progresso

Cada API documenta sua tabela de imports no README local. Use as consultas sugeridas em cada modulo para acompanhar:

- status;
- ultima etapa executada;
- quantidade de linhas inseridas;
- erro registrado.

## Observacoes importantes

- as APIs sao independentes: voce nao precisa executar todas;
- o banco de destino e seu;
- os modulos foram feitos para permitir reexecucao e retomada;
- antes de abrir issue, rode exatamente os comandos do README local do modulo.
