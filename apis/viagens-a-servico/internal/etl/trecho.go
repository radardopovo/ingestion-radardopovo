// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"time"

	"github.com/radardopovo/viagens-etl/internal/db"
	"github.com/radardopovo/viagens-etl/internal/parse"
)

func (o *Orchestrator) importTrechos(ctx context.Context, year int, csvPath string) (TableStats, error) {
	spec := db.BulkTableSpec{
		Table: "trechos",
		Columns: []string{
			"id",
			"processo_id",
			"pcdp",
			"sequencia",
			"origem_data",
			"origem_pais",
			"origem_uf",
			"origem_cidade",
			"destino_data",
			"destino_pais",
			"destino_uf",
			"destino_cidade",
			"meio_transporte",
			"numero_diarias",
			"missao",
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

	return o.runCSVImport(ctx, year, "Trecho", csvPath, spec, required, func(get fieldGetter) ([]any, bool) {
		processoID := get("Identificador do processo de viagem")
		pcdp := get("Numero da Proposta (PCDP)")
		if processoID == "" || pcdp == "" {
			return nil, false
		}

		sequenciaRaw := get("Sequencia Trecho")
		var origemData any
		if t, ok := parse.DateBR(get("Origem - Data")); ok {
			origemData = t
		}
		var destinoData any
		if t, ok := parse.DateBR(get("Destino - Data")); ok {
			destinoData = t
		}

		id := parse.MakeID(
			processoID,
			pcdp,
			sequenciaRaw,
			get("Origem - Data"),
			get("Origem - Cidade"),
			get("Destino - Data"),
			get("Destino - Cidade"),
			get("Meio de transporte"),
		)

		importedAt := time.Now().UTC()
		return []any{
			id,
			processoID,
			pcdp,
			intOrNil(sequenciaRaw),
			origemData,
			nullableString(get("Origem - Pais")),
			nullableString(get("Origem - UF")),
			nullableString(get("Origem - Cidade")),
			destinoData,
			nullableString(get("Destino - Pais")),
			nullableString(get("Destino - UF")),
			nullableString(get("Destino - Cidade")),
			nullableString(get("Meio de transporte")),
			decimalOrNil(get("Numero Diarias")),
			nullableString(get("Missao?")),
			year,
			importedAt,
		}, true
	})
}
