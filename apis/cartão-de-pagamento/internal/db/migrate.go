// Radar do Povo ETL - https://radardopovo.com
package db

import (
	"context"
	"database/sql"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS cpgf (
			id                      CHAR(40)    PRIMARY KEY,
			codigo_orgao_superior   VARCHAR(30),
			nome_orgao_superior     VARCHAR(255),
			codigo_orgao            VARCHAR(30),
			nome_orgao              VARCHAR(255),
			codigo_unidade_gestora  VARCHAR(30),
			nome_unidade_gestora    VARCHAR(255),
			ano_extrato             SMALLINT,
			mes_extrato             SMALLINT,
			cpf_portador            VARCHAR(20),
			nome_portador           VARCHAR(255),
			documento_favorecido    VARCHAR(20),
			nome_favorecido         TEXT,
			transacao               TEXT,
			data_transacao          DATE,
			valor_transacao_cents   BIGINT,
			periodo_yyyymm          INT         NOT NULL,
			imported_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cpgf_periodo ON cpgf (periodo_yyyymm)`,
		`CREATE INDEX IF NOT EXISTS idx_cpgf_data_transacao ON cpgf (data_transacao)`,
		`CREATE INDEX IF NOT EXISTS idx_cpgf_orgao_superior ON cpgf (codigo_orgao_superior)`,
		`CREATE INDEX IF NOT EXISTS idx_cpgf_doc_favorecido ON cpgf (documento_favorecido)`,
		`CREATE TABLE IF NOT EXISTS imports_cpgf (
			periodo_yyyymm  INT         PRIMARY KEY,
			zip_sha256      CHAR(64)    NOT NULL,
			status          VARCHAR(20) NOT NULL CHECK (status IN ('downloading','extracted','importing','done','error')),
			last_step       VARCHAR(20) CHECK (last_step IN ('download','extract','cpgf')),
			rows_cpgf       BIGINT DEFAULT 0,
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
