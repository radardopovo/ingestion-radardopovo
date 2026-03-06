// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"time"

	"github.com/radardopovo/emendas-etl/internal/db"
	"github.com/radardopovo/emendas-etl/internal/parse"
)

func (o *Orchestrator) importEmendasConvenios(ctx context.Context, csvPath string) (TableStats, error) {
	spec := db.BulkTableSpec{
		Table: "emendas_convenios",
		Columns: []string{
			"id",
			"codigo_emenda",
			"codigo_funcao",
			"nome_funcao",
			"codigo_subfuncao",
			"nome_subfuncao",
			"localidade_gasto",
			"tipo_emenda",
			"data_publicacao_convenio",
			"convenente",
			"objeto_convenio",
			"numero_convenio",
			"valor_convenio_cents",
			"imported_at",
		},
		ConflictTarget:          "id",
		UpsertColumns:           nil,
		CountConflictsAsIgnored: true,
	}

	required := []string{"Codigo da Emenda", "Numero Convenio"}

	return o.runCSVImport(ctx, datasetKey, "Convenios", csvPath, spec, required, func(get fieldGetter) ([]any, bool) {
		codigoEmenda := get("Codigo da Emenda")
		numeroConvenio := get("Numero Convenio")
		if codigoEmenda == "" || numeroConvenio == "" {
			return nil, false
		}

		id := parse.MakeID(
			codigoEmenda,
			numeroConvenio,
			get("Data Publicacao Convenio"),
			get("Valor Convenio"),
		)

		var dataPublicacao any
		if t, ok := parse.DateBR(get("Data Publicacao Convenio")); ok {
			dataPublicacao = t
		}
		var valorConvenio any
		if v, ok := parse.MoneyToCents(get("Valor Convenio")); ok {
			valorConvenio = v
		}

		importedAt := time.Now().UTC()
		return []any{
			id,
			codigoEmenda,
			nullableString(get("Codigo Funcao")),
			nullableString(get("Nome Funcao")),
			nullableString(get("Codigo Subfuncao")),
			nullableString(get("Nome Subfuncao")),
			nullableString(get("Localidade do gasto")),
			nullableString(get("Tipo de Emenda")),
			dataPublicacao,
			nullableString(get("Convenente")),
			nullableString(get("Objeto Convenio")),
			nullableString(numeroConvenio),
			valorConvenio,
			importedAt,
		}, true
	})
}
