// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"time"

	"github.com/radardopovo/emendas-etl/internal/db"
	"github.com/radardopovo/emendas-etl/internal/parse"
)

func (o *Orchestrator) importEmendas(ctx context.Context, csvPath string) (TableStats, error) {
	spec := db.BulkTableSpec{
		Table: "emendas",
		Columns: []string{
			"id",
			"codigo_emenda",
			"ano_emenda",
			"tipo_emenda",
			"codigo_autor_emenda",
			"nome_autor_emenda",
			"numero_emenda",
			"localidade_aplicacao",
			"codigo_municipio_ibge",
			"municipio",
			"codigo_uf_ibge",
			"uf",
			"regiao",
			"codigo_funcao",
			"nome_funcao",
			"codigo_subfuncao",
			"nome_subfuncao",
			"codigo_programa",
			"nome_programa",
			"codigo_acao",
			"nome_acao",
			"codigo_plano_orcamentario",
			"nome_plano_orcamentario",
			"valor_empenhado_cents",
			"valor_liquidado_cents",
			"valor_pago_cents",
			"valor_rp_inscritos_cents",
			"valor_rp_cancelados_cents",
			"valor_rp_pagos_cents",
			"imported_at",
		},
		ConflictTarget:          "id",
		UpsertColumns:           nil,
		CountConflictsAsIgnored: true,
	}

	required := []string{"Codigo da Emenda"}

	return o.runCSVImport(ctx, datasetKey, "Emendas", csvPath, spec, required, func(get fieldGetter) ([]any, bool) {
		codigoEmenda := get("Codigo da Emenda")
		if codigoEmenda == "" {
			return nil, false
		}

		id := parse.MakeID(
			codigoEmenda,
			get("Ano da Emenda"),
			get("Codigo do Autor da Emenda"),
			get("Numero da emenda"),
			get("Codigo Municipio IBGE"),
			get("Codigo Funcao"),
			get("Codigo Subfuncao"),
			get("Codigo Programa"),
			get("Codigo Acao"),
			get("Codigo Plano Orcamentario"),
		)

		var valorEmpenhado any
		if v, ok := parse.MoneyToCents(get("Valor Empenhado")); ok {
			valorEmpenhado = v
		}
		var valorLiquidado any
		if v, ok := parse.MoneyToCents(get("Valor Liquidado")); ok {
			valorLiquidado = v
		}
		var valorPago any
		if v, ok := parse.MoneyToCents(get("Valor Pago")); ok {
			valorPago = v
		}
		var valorRPInscritos any
		if v, ok := parse.MoneyToCents(get("Valor Restos A Pagar Inscritos")); ok {
			valorRPInscritos = v
		}
		var valorRPCancelados any
		if v, ok := parse.MoneyToCents(get("Valor Restos A Pagar Cancelados")); ok {
			valorRPCancelados = v
		}
		var valorRPPagos any
		if v, ok := parse.MoneyToCents(get("Valor Restos A Pagar Pagos")); ok {
			valorRPPagos = v
		}

		importedAt := time.Now().UTC()
		return []any{
			id,
			codigoEmenda,
			intOrNil(get("Ano da Emenda")),
			nullableString(get("Tipo de Emenda")),
			nullableString(get("Codigo do Autor da Emenda")),
			nullableString(get("Nome do Autor da Emenda")),
			nullableString(get("Numero da emenda")),
			nullableString(get("Localidade de aplicacao do recurso")),
			nullableString(get("Codigo Municipio IBGE")),
			nullableString(get("Municipio")),
			nullableString(get("Codigo UF IBGE")),
			nullableString(get("UF")),
			nullableString(get("Regiao")),
			nullableString(get("Codigo Funcao")),
			nullableString(get("Nome Funcao")),
			nullableString(get("Codigo Subfuncao")),
			nullableString(get("Nome Subfuncao")),
			nullableString(get("Codigo Programa")),
			nullableString(get("Nome Programa")),
			nullableString(get("Codigo Acao")),
			nullableString(get("Nome Acao")),
			nullableString(get("Codigo Plano Orcamentario")),
			nullableString(get("Nome Plano Orcamentario")),
			valorEmpenhado,
			valorLiquidado,
			valorPago,
			valorRPInscritos,
			valorRPCancelados,
			valorRPPagos,
			importedAt,
		}, true
	})
}
