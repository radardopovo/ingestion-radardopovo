// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"strconv"
	"time"

	"github.com/radardopovo/cpgf-etl/internal/db"
	"github.com/radardopovo/cpgf-etl/internal/parse"
)

func (o *Orchestrator) importCPGF(ctx context.Context, period int, csvPath string) (TableStats, error) {
	spec := db.BulkTableSpec{
		Table: "cpgf",
		Columns: []string{
			"id",
			"codigo_orgao_superior",
			"nome_orgao_superior",
			"codigo_orgao",
			"nome_orgao",
			"codigo_unidade_gestora",
			"nome_unidade_gestora",
			"ano_extrato",
			"mes_extrato",
			"cpf_portador",
			"nome_portador",
			"documento_favorecido",
			"nome_favorecido",
			"transacao",
			"data_transacao",
			"valor_transacao_cents",
			"periodo_yyyymm",
			"imported_at",
		},
		ConflictTarget:          "id",
		UpsertColumns:           nil,
		CountConflictsAsIgnored: true,
	}

	required := []string{
		"CODIGO ORGAO SUPERIOR",
		"DATA TRANSACAO",
		"VALOR TRANSACAO",
	}

	return o.runCSVImport(ctx, period, "CPGF", csvPath, spec, required, func(get fieldGetter) ([]any, bool) {
		codigoOrgaoSuperior := get("CODIGO ORGAO SUPERIOR")
		nomeOrgaoSuperior := get("NOME ORGAO SUPERIOR")
		codigoOrgao := get("CODIGO ORGAO")
		nomeOrgao := get("NOME ORGAO")
		codigoUG := get("CODIGO UNIDADE GESTORA")
		nomeUG := get("NOME UNIDADE GESTORA")
		anoExtratoRaw := get("ANO EXTRATO")
		mesExtratoRaw := get("MES EXTRATO")
		cpfPortador := get("CPF PORTADOR")
		nomePortador := get("NOME PORTADOR")
		docFavorecido := get("CNPJ OU CPF FAVORECIDO")
		nomeFavorecido := get("NOME FAVORECIDO")
		transacao := get("TRANSACAO")
		dataTransacaoRaw := get("DATA TRANSACAO")
		valorTransacaoRaw := get("VALOR TRANSACAO")

		if isBlankRecord(codigoOrgaoSuperior, codigoOrgao, codigoUG, cpfPortador, docFavorecido, transacao, dataTransacaoRaw, valorTransacaoRaw) {
			return nil, false
		}

		periodYear := period / 100
		periodMonth := period % 100

		anoExtrato := parseIntOrFallback(anoExtratoRaw, periodYear)
		mesExtrato := parseIntOrFallback(mesExtratoRaw, periodMonth)

		var dataTransacao any
		if dt, ok := parse.DateBR(dataTransacaoRaw); ok {
			dataTransacao = dt
		}

		var valorTransacao any
		if v, ok := parse.MoneyToCents(valorTransacaoRaw); ok {
			valorTransacao = v
		}

		id := parse.MakeID(
			strconv.Itoa(period),
			codigoOrgaoSuperior,
			nomeOrgaoSuperior,
			codigoOrgao,
			nomeOrgao,
			codigoUG,
			nomeUG,
			anoExtratoRaw,
			mesExtratoRaw,
			cpfPortador,
			nomePortador,
			docFavorecido,
			nomeFavorecido,
			transacao,
			dataTransacaoRaw,
			valorTransacaoRaw,
		)

		importedAt := time.Now().UTC()
		return []any{
			id,
			nullableString(codigoOrgaoSuperior),
			nullableString(nomeOrgaoSuperior),
			nullableString(codigoOrgao),
			nullableString(nomeOrgao),
			nullableString(codigoUG),
			nullableString(nomeUG),
			anoExtrato,
			mesExtrato,
			nullableString(cpfPortador),
			nullableString(nomePortador),
			nullableString(docFavorecido),
			nullableString(nomeFavorecido),
			nullableString(transacao),
			dataTransacao,
			valorTransacao,
			period,
			importedAt,
		}, true
	})
}

func parseIntOrFallback(raw string, fallback int) any {
	v := cleanString(raw)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func isBlankRecord(values ...string) bool {
	for _, v := range values {
		if cleanString(v) != "" {
			return false
		}
	}
	return true
}
