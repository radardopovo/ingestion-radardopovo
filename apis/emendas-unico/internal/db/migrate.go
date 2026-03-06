// Radar do Povo ETL - https://radardopovo.com
package db

import (
	"context"
	"database/sql"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS emendas (
			id                               CHAR(40) PRIMARY KEY,
			codigo_emenda                    VARCHAR(80),
			ano_emenda                       SMALLINT,
			tipo_emenda                      TEXT,
			codigo_autor_emenda              VARCHAR(80),
			nome_autor_emenda                TEXT,
			numero_emenda                    VARCHAR(80),
			localidade_aplicacao             TEXT,
			codigo_municipio_ibge            VARCHAR(20),
			municipio                        VARCHAR(255),
			codigo_uf_ibge                   VARCHAR(20),
			uf                               VARCHAR(120),
			regiao                           VARCHAR(120),
			codigo_funcao                    VARCHAR(20),
			nome_funcao                      VARCHAR(255),
			codigo_subfuncao                 VARCHAR(20),
			nome_subfuncao                   VARCHAR(255),
			codigo_programa                  VARCHAR(20),
			nome_programa                    TEXT,
			codigo_acao                      VARCHAR(20),
			nome_acao                        TEXT,
			codigo_plano_orcamentario        VARCHAR(30),
			nome_plano_orcamentario          TEXT,
			valor_empenhado_cents            BIGINT,
			valor_liquidado_cents            BIGINT,
			valor_pago_cents                 BIGINT,
			valor_rp_inscritos_cents         BIGINT,
			valor_rp_cancelados_cents        BIGINT,
			valor_rp_pagos_cents             BIGINT,
			imported_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_emendas_codigo_emenda ON emendas (codigo_emenda)`,
		`CREATE INDEX IF NOT EXISTS idx_emendas_ano_emenda ON emendas (ano_emenda)`,
		`CREATE INDEX IF NOT EXISTS idx_emendas_codigo_autor ON emendas (codigo_autor_emenda)`,
		`CREATE TABLE IF NOT EXISTS emendas_por_favorecido (
			id                               CHAR(40) PRIMARY KEY,
			codigo_emenda                    VARCHAR(80),
			codigo_autor_emenda              VARCHAR(80),
			nome_autor_emenda                TEXT,
			numero_emenda                    VARCHAR(80),
			tipo_emenda                      TEXT,
			ano_mes                          VARCHAR(20),
			codigo_favorecido                VARCHAR(80),
			favorecido                       TEXT,
			natureza_juridica                TEXT,
			tipo_favorecido                  VARCHAR(120),
			uf_favorecido                    VARCHAR(80),
			municipio_favorecido             VARCHAR(255),
			valor_recebido_cents             BIGINT,
			imported_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_favorecido_codigo_emenda ON emendas_por_favorecido (codigo_emenda)`,
		`CREATE INDEX IF NOT EXISTS idx_favorecido_ano_mes ON emendas_por_favorecido (ano_mes)`,
		`CREATE INDEX IF NOT EXISTS idx_favorecido_codigo_fav ON emendas_por_favorecido (codigo_favorecido)`,
		`CREATE TABLE IF NOT EXISTS emendas_convenios (
			id                               CHAR(40) PRIMARY KEY,
			codigo_emenda                    VARCHAR(80),
			codigo_funcao                    VARCHAR(20),
			nome_funcao                      VARCHAR(255),
			codigo_subfuncao                 VARCHAR(20),
			nome_subfuncao                   VARCHAR(255),
			localidade_gasto                 TEXT,
			tipo_emenda                      TEXT,
			data_publicacao_convenio         DATE,
			convenente                       TEXT,
			objeto_convenio                  TEXT,
			numero_convenio                  VARCHAR(80),
			valor_convenio_cents             BIGINT,
			imported_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_convenios_codigo_emenda ON emendas_convenios (codigo_emenda)`,
		`CREATE INDEX IF NOT EXISTS idx_convenios_data_pub ON emendas_convenios (data_publicacao_convenio)`,
		`DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'imports'
			) AND EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'imports' AND column_name = 'ano'
			) THEN
				DROP TABLE imports;
			END IF;
		END $$`,
		`CREATE TABLE IF NOT EXISTS imports (
			dataset_key         VARCHAR(80) PRIMARY KEY,
			zip_sha256          CHAR(64)    NOT NULL,
			status              VARCHAR(20) NOT NULL CHECK (status IN ('downloading','extracted','importing','done','error')),
			last_step           VARCHAR(20) CHECK (last_step IN ('download','extract','emendas','favorecido','convenio')),
			rows_emendas        BIGINT DEFAULT 0,
			rows_favorecido     BIGINT DEFAULT 0,
			rows_convenio       BIGINT DEFAULT 0,
			started_at          TIMESTAMPTZ,
			finished_at         TIMESTAMPTZ,
			error_msg           TEXT
		)`,
		`ALTER TABLE imports ADD COLUMN IF NOT EXISTS zip_sha256 CHAR(64)`,
		`ALTER TABLE imports ADD COLUMN IF NOT EXISTS status VARCHAR(20)`,
		`ALTER TABLE imports ADD COLUMN IF NOT EXISTS last_step VARCHAR(20)`,
		`ALTER TABLE imports ADD COLUMN IF NOT EXISTS rows_emendas BIGINT DEFAULT 0`,
		`ALTER TABLE imports ADD COLUMN IF NOT EXISTS rows_favorecido BIGINT DEFAULT 0`,
		`ALTER TABLE imports ADD COLUMN IF NOT EXISTS rows_convenio BIGINT DEFAULT 0`,
		`ALTER TABLE imports ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ`,
		`ALTER TABLE imports ADD COLUMN IF NOT EXISTS finished_at TIMESTAMPTZ`,
		`ALTER TABLE imports ADD COLUMN IF NOT EXISTS error_msg TEXT`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
