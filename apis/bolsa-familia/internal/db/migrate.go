// Radar do Povo ETL - https://radardopovo.com
package db

import (
	"context"
	"database/sql"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS bolsa_familia (
			id                      CHAR(40)    PRIMARY KEY,
			mes_competencia         INT,
			mes_referencia          INT,
			uf                      VARCHAR(10),
			codigo_municipio_siafi  VARCHAR(20),
			nome_municipio          VARCHAR(255),
			cpf_favorecido          VARCHAR(20),
			nis_favorecido          VARCHAR(20),
			nome_favorecido         VARCHAR(255),
			valor_parcela_cents     BIGINT,
			periodo_yyyymm          INT         NOT NULL,
			imported_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_bolsa_familia_periodo ON bolsa_familia (periodo_yyyymm)`,
		`CREATE INDEX IF NOT EXISTS idx_bolsa_familia_uf ON bolsa_familia (uf)`,
		`CREATE INDEX IF NOT EXISTS idx_bolsa_familia_municipio ON bolsa_familia (codigo_municipio_siafi)`,
		`CREATE INDEX IF NOT EXISTS idx_bolsa_familia_nis ON bolsa_familia (nis_favorecido)`,
		`CREATE TABLE IF NOT EXISTS imports_bolsa_familia (
			periodo_yyyymm  INT         PRIMARY KEY,
			zip_sha256      CHAR(64)    NOT NULL,
			status          VARCHAR(20) NOT NULL CHECK (status IN ('downloading','extracted','importing','done','error')),
			last_step       VARCHAR(20) CHECK (last_step IN ('download','extract','bolsa_familia')),
			rows_bolsa_familia BIGINT DEFAULT 0,
			started_at      TIMESTAMPTZ,
			finished_at     TIMESTAMPTZ,
			error_msg       TEXT
		)`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
