// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"time"

	"github.com/radardopovo/viagens-etl/internal/db"
	"github.com/radardopovo/viagens-etl/internal/parse"
)

func (o *Orchestrator) importViagens(ctx context.Context, year int, csvPath string) (TableStats, error) {
	spec := db.BulkTableSpec{
		Table: "viagens",
		Columns: []string{
			"processo_id",
			"pcdp",
			"situacao",
			"viagem_urgente",
			"justificativa_urgencia",
			"orgao_superior_codigo",
			"orgao_superior_nome",
			"orgao_solicitante_codigo",
			"orgao_solicitante_nome",
			"cpf_viajante",
			"nome_viajante",
			"cargo",
			"funcao",
			"descricao_funcao",
			"data_inicio",
			"data_fim",
			"destinos",
			"motivo",
			"valor_diarias_cents",
			"valor_passagens_cents",
			"valor_devolucao_cents",
			"valor_outros_gastos_cents",
			"ano",
			"imported_at",
		},
		ConflictTarget:          "processo_id",
		UpsertColumns:           nil,
		CountConflictsAsIgnored: true,
	}

	required := []string{
		"Identificador do processo de viagem",
		"Numero da Proposta (PCDP)",
	}

	return o.runCSVImport(ctx, year, "Viagem", csvPath, spec, required, func(get fieldGetter) ([]any, bool) {
		processoID := get("Identificador do processo de viagem")
		pcdp := get("Numero da Proposta (PCDP)")
		if processoID == "" || pcdp == "" {
			return nil, false
		}

		var dataInicio any
		if t, ok := parse.DateBR(get("Periodo - Data de inicio")); ok {
			dataInicio = t
		}
		var dataFim any
		if t, ok := parse.DateBR(get("Periodo - Data de fim")); ok {
			dataFim = t
		}
		var valorDiarias any
		if v, ok := parse.MoneyToCents(get("Valor diarias")); ok {
			valorDiarias = v
		}
		var valorPassagens any
		if v, ok := parse.MoneyToCents(get("Valor passagens")); ok {
			valorPassagens = v
		}
		var valorDevolucao any
		if v, ok := parse.MoneyToCents(get("Valor devolucao")); ok {
			valorDevolucao = v
		}
		var valorOutros any
		if v, ok := parse.MoneyToCents(get("Valor outros gastos")); ok {
			valorOutros = v
		}

		importedAt := time.Now().UTC()
		return []any{
			processoID,
			pcdp,
			nullableString(get("Situacao")),
			nullableString(get("Viagem Urgente")),
			nullableString(get("Justificativa Urgencia Viagem")),
			nullableString(get("Codigo do orgao superior")),
			nullableString(get("Nome do orgao superior")),
			nullableString(get("Codigo orgao solicitante")),
			nullableString(get("Nome orgao solicitante")),
			nullableString(get("CPF viajante")),
			nullableString(get("Nome")),
			nullableString(get("Cargo")),
			nullableString(get("Funcao")),
			nullableString(get("Descricao Funcao")),
			dataInicio,
			dataFim,
			nullableString(get("Destinos")),
			nullableString(get("Motivo")),
			valorDiarias,
			valorPassagens,
			valorDevolucao,
			valorOutros,
			year,
			importedAt,
		}, true
	})
}
