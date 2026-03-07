// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"time"

	"github.com/radardopovo/viagens-etl/internal/db"
	"github.com/radardopovo/viagens-etl/internal/parse"
)

func (o *Orchestrator) importPagamentos(ctx context.Context, year int, csvPath string) (TableStats, error) {
	spec := db.BulkTableSpec{
		Table: "pagamentos",
		Columns: []string{
			"id",
			"processo_id",
			"pcdp",
			"orgao_superior_codigo",
			"orgao_superior_nome",
			"orgao_pagador_codigo",
			"orgao_pagador_nome",
			"ug_pagadora_codigo",
			"ug_pagadora_nome",
			"tipo_pagamento",
			"valor_cents",
			"ano",
			"imported_at",
		},
		ConflictTarget:          "id",
		UpsertColumns:           nil,
		CountConflictsAsIgnored: true,
	}

	required := []string{
		"Identificador do processo de viagem",
		"Numero da Proposta (PCDP)",
	}

	return o.runCSVImport(ctx, year, "Pagamento", csvPath, spec, required, func(get fieldGetter) ([]any, bool) {
		processoID := get("Identificador do processo de viagem")
		pcdp := get("Numero da Proposta (PCDP)")
		if processoID == "" || pcdp == "" {
			return nil, false
		}

		var valor any
		if v, ok := parse.MoneyToCents(get("Valor")); ok {
			valor = v
		}

		id := parse.MakeID(
			processoID,
			pcdp,
			get("Tipo de pagamento"),
			get("Valor"),
			get("Codigo do orgao pagador"),
			get("Codigo da unidade gestora pagadora"),
		)

		importedAt := time.Now().UTC()
		return []any{
			id,
			processoID,
			pcdp,
			nullableString(get("Codigo do orgao superior")),
			nullableString(get("Nome do orgao superior")),
			nullableString(get("Codigo do orgao pagador")),
			nullableString(get("Nome do orgao pagador")),
			nullableString(get("Codigo da unidade gestora pagadora")),
			nullableString(get("Nome da unidade gestora pagadora")),
			nullableString(get("Tipo de pagamento")),
			valor,
			year,
			importedAt,
		}, true
	})
}
