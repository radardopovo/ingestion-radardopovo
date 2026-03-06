// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"time"

	"github.com/radardopovo/emendas-etl/internal/db"
	"github.com/radardopovo/emendas-etl/internal/parse"
)

func (o *Orchestrator) importEmendasPorFavorecido(ctx context.Context, csvPath string) (TableStats, error) {
	spec := db.BulkTableSpec{
		Table: "emendas_por_favorecido",
		Columns: []string{
			"id",
			"codigo_emenda",
			"codigo_autor_emenda",
			"nome_autor_emenda",
			"numero_emenda",
			"tipo_emenda",
			"ano_mes",
			"codigo_favorecido",
			"favorecido",
			"natureza_juridica",
			"tipo_favorecido",
			"uf_favorecido",
			"municipio_favorecido",
			"valor_recebido_cents",
			"imported_at",
		},
		ConflictTarget:          "id",
		UpsertColumns:           nil,
		CountConflictsAsIgnored: true,
	}

	required := []string{"Codigo da Emenda", "Ano/Mes"}

	return o.runCSVImport(ctx, datasetKey, "PorFavorecido", csvPath, spec, required, func(get fieldGetter) ([]any, bool) {
		codigoEmenda := get("Codigo da Emenda")
		anoMes := get("Ano/Mes")
		if codigoEmenda == "" || anoMes == "" {
			return nil, false
		}

		id := parse.MakeID(
			codigoEmenda,
			get("Codigo do Autor da Emenda"),
			get("Numero da emenda"),
			anoMes,
			get("Codigo do Favorecido"),
			get("Valor Recebido"),
		)

		var valorRecebido any
		if v, ok := parse.MoneyToCents(get("Valor Recebido")); ok {
			valorRecebido = v
		}

		importedAt := time.Now().UTC()
		return []any{
			id,
			codigoEmenda,
			nullableString(get("Codigo do Autor da Emenda")),
			nullableString(get("Nome do Autor da Emenda")),
			nullableString(get("Numero da emenda")),
			nullableString(get("Tipo de Emenda")),
			nullableString(anoMes),
			nullableString(get("Codigo do Favorecido")),
			nullableString(get("Favorecido")),
			nullableString(get("Natureza Juridica")),
			nullableString(get("Tipo Favorecido")),
			nullableString(get("UF Favorecido")),
			nullableString(get("Municipio Favorecido")),
			valorRecebido,
			importedAt,
		}, true
	})
}
