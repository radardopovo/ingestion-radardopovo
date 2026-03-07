// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"strconv"
	"time"

	"github.com/radardopovo/bolsa-familia-etl/internal/db"
	"github.com/radardopovo/bolsa-familia-etl/internal/parse"
)

func (o *Orchestrator) importBolsaFamilia(ctx context.Context, period int, csvPath string) (TableStats, error) {
	spec := db.BulkTableSpec{
		Table: "bolsa_familia",
		Columns: []string{
			"id",
			"mes_competencia",
			"mes_referencia",
			"uf",
			"codigo_municipio_siafi",
			"nome_municipio",
			"cpf_favorecido",
			"nis_favorecido",
			"nome_favorecido",
			"valor_parcela_cents",
			"periodo_yyyymm",
			"imported_at",
		},
		ConflictTarget:          "id",
		UpsertColumns:           nil,
		CountConflictsAsIgnored: true,
	}

	required := []string{
		"MES COMPETENCIA",
		"MES REFERENCIA",
		"UF",
		"CODIGO MUNICIPIO SIAFI",
		"NOME MUNICIPIO",
		"NIS FAVORECIDO",
		"NOME FAVORECIDO",
		"VALOR PARCELA",
	}

	return o.runCSVImport(ctx, period, "NovoBolsaFamilia", csvPath, spec, required, func(get fieldGetter) ([]any, bool) {
		mesCompetenciaRaw := get("MES COMPETENCIA")
		mesReferenciaRaw := get("MES REFERENCIA")
		uf := get("UF")
		codigoMunicipio := get("CODIGO MUNICIPIO SIAFI")
		nomeMunicipio := get("NOME MUNICIPIO")
		cpfFavorecido := get("CPF FAVORECIDO")
		nisFavorecido := get("NIS FAVORECIDO")
		nomeFavorecido := get("NOME FAVORECIDO")
		valorParcelaRaw := get("VALOR PARCELA")

		if isBlankRecord(mesCompetenciaRaw, mesReferenciaRaw, uf, codigoMunicipio, nomeMunicipio, cpfFavorecido, nisFavorecido, nomeFavorecido, valorParcelaRaw) {
			return nil, false
		}

		mesCompetencia := parseIntOrFallback(mesCompetenciaRaw, period)
		mesReferencia := parseIntOrFallback(mesReferenciaRaw, period)

		var valorParcela any
		if v, ok := parse.MoneyToCents(valorParcelaRaw); ok {
			valorParcela = v
		}

		id := parse.MakeID(
			strconv.Itoa(period),
			mesCompetenciaRaw,
			mesReferenciaRaw,
			uf,
			codigoMunicipio,
			nomeMunicipio,
			cpfFavorecido,
			nisFavorecido,
			nomeFavorecido,
			valorParcelaRaw,
		)

		importedAt := time.Now().UTC()
		return []any{
			id,
			mesCompetencia,
			mesReferencia,
			nullableString(uf),
			nullableString(codigoMunicipio),
			nullableString(nomeMunicipio),
			nullableString(cpfFavorecido),
			nullableString(nisFavorecido),
			nullableString(nomeFavorecido),
			valorParcela,
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
