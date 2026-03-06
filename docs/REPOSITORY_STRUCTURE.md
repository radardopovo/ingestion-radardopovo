# Estrutura do repositório

O Radar do Povo adota um monorepo orientado por dataset.

## Regras

1. cada dataset vive em `apis/<dataset-slug>`;
2. datasets compartilham a mesma filosofia estrutural;
3. schemas SQL e docs ficam próximos ao dataset;
4. componentes comuns podem surgir futuramente em `pkg/` ou `internal/shared/`, mas sem antecipar abstrações desnecessárias.
