// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"time"

	"github.com/radardopovo/viagens-etl/internal/db"
	"github.com/radardopovo/viagens-etl/internal/parse"
)

func (o *Orchestrator) importPassagens(ctx context.Context, year int, csvPath string) (TableStats, error) {
	spec := db.BulkTableSpec{
		Table: "passagens",
		Columns: []string{
			"id",
			"processo_id",
			"pcdp",
			"meio_transporte",
			"ida_origem_pais",
			"ida_origem_uf",
			"ida_origem_cidade",
			"ida_destino_pais",
			"ida_destino_uf",
			"ida_destino_cidade",
			"volta_origem_pais",
			"volta_origem_uf",
			"volta_origem_cidade",
			"volta_destino_pais",
			"volta_destino_uf",
			"volta_destino_cidade",
			"valor_passagem_cents",
			"taxa_servico_cents",
			"emissao_data",
			"emissao_hora",
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

	return o.runCSVImport(ctx, year, "Passagem", csvPath, spec, required, func(get fieldGetter) ([]any, bool) {
		processoID := get("Identificador do processo de viagem")
		pcdp := get("Numero da Proposta (PCDP)")
		if processoID == "" || pcdp == "" {
			return nil, false
		}

		var emissaoData any
		if t, ok := parse.DateBR(get("Data da emissao/compra")); ok {
			emissaoData = t
		}
		var valorPassagem any
		if v, ok := parse.MoneyToCents(get("Valor da passagem")); ok {
			valorPassagem = v
		}
		var taxaServico any
		if v, ok := parse.MoneyToCents(get("Taxa de servico")); ok {
			taxaServico = v
		}

		id := parse.MakeID(
			processoID,
			pcdp,
			get("Meio de transporte"),
			get("Pais - Origem ida"),
			get("Cidade - Origem ida"),
			get("Pais - Destino ida"),
			get("Cidade - Destino ida"),
			get("Pais - Origem volta"),
			get("Cidade - Origem volta"),
			get("Pais - Destino volta"),
			get("Cidade - Destino volta"),
			get("Data da emissao/compra"),
			get("Hora da emissao/compra"),
			get("Valor da passagem"),
			get("Taxa de servico"),
		)

		importedAt := time.Now().UTC()
		return []any{
			id,
			processoID,
			pcdp,
			nullableString(get("Meio de transporte")),
			nullableString(get("Pais - Origem ida")),
			nullableString(get("UF - Origem ida")),
			nullableString(get("Cidade - Origem ida")),
			nullableString(get("Pais - Destino ida")),
			nullableString(get("UF - Destino ida")),
			nullableString(get("Cidade - Destino ida")),
			nullableString(get("Pais - Origem volta")),
			nullableString(get("UF - Origem volta")),
			nullableString(get("Cidade - Origem volta")),
			nullableString(get("Pais - Destino volta")),
			nullableString(get("UF - Destino volta")),
			nullableString(get("Cidade - Destino volta")),
			valorPassagem,
			taxaServico,
			emissaoData,
			nullableString(get("Hora da emissao/compra")),
			year,
			importedAt,
		}, true
	})
}
