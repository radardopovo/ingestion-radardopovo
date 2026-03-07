# Contributing

Projeto mantido por Radar do Povo: [radardopovo.com](https://radardopovo.com)

## Requisitos

- Go 1.23+
- Acesso ao PostgreSQL com TLS `verify-full`

## Fluxo recomendado

1. Crie uma branch para a alteracao.
2. Rode validacao local:
   - `go test ./...`
   - `go vet ./...`
   - `go build ./...`
3. Atualize README se houver mudanca de comportamento.
4. Abra PR com descricao objetiva, riscos e plano de rollback.

## Padroes do repositorio

- Sem ORM, sem dependencias fora de `github.com/lib/pq` e `golang.org/x/text`.
- Importador precisa manter idempotencia e resume.
- Logs devem ser estruturados e conter contexto suficiente para diagnostico.
