// Radar do Povo ETL - https://radardopovo.com
package db

import (
	"context"
	"database/sql"
)

func Migrate(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS viagens (
			processo_id               VARCHAR(80)  PRIMARY KEY,
			pcdp                      VARCHAR(80)  NOT NULL,
			situacao                  VARCHAR(120),
			viagem_urgente            VARCHAR(50),
			justificativa_urgencia    TEXT,
			orgao_superior_codigo     VARCHAR(30),
			orgao_superior_nome       VARCHAR(255),
			orgao_solicitante_codigo  VARCHAR(30),
			orgao_solicitante_nome    VARCHAR(255),
			cpf_viajante              VARCHAR(20),
			nome_viajante             VARCHAR(255),
			cargo                     VARCHAR(255),
			funcao                    VARCHAR(255),
			descricao_funcao          TEXT,
			data_inicio               DATE,
			data_fim                  DATE,
			destinos                  TEXT,
			motivo                    TEXT,
			valor_diarias_cents       BIGINT,
			valor_passagens_cents     BIGINT,
			valor_devolucao_cents     BIGINT,
			valor_outros_gastos_cents BIGINT,
			ano                       SMALLINT     NOT NULL,
			imported_at               TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_viagens_ano ON viagens (ano)`,
		`CREATE INDEX IF NOT EXISTS idx_viagens_orgao_superior_codigo ON viagens (orgao_superior_codigo)`,
		`CREATE INDEX IF NOT EXISTS idx_viagens_orgao_solicitante_codigo ON viagens (orgao_solicitante_codigo)`,
		`CREATE INDEX IF NOT EXISTS idx_viagens_nome_viajante ON viagens (nome_viajante)`,
		`CREATE INDEX IF NOT EXISTS idx_viagens_data_inicio ON viagens (data_inicio)`,
		`CREATE TABLE IF NOT EXISTS trechos (
			id               CHAR(40)    PRIMARY KEY,
			processo_id      VARCHAR(80) NOT NULL,
			pcdp             VARCHAR(80) NOT NULL,
			sequencia        INT,
			origem_data      DATE,
			origem_pais      VARCHAR(120),
			origem_uf        VARCHAR(40),
			origem_cidade    VARCHAR(255),
			destino_data     DATE,
			destino_pais     VARCHAR(120),
			destino_uf       VARCHAR(40),
			destino_cidade   VARCHAR(255),
			meio_transporte  VARCHAR(120),
			numero_diarias   NUMERIC(10,2),
			missao           VARCHAR(255),
			ano              SMALLINT    NOT NULL,
			imported_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_trechos_processo ON trechos (processo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_trechos_ano ON trechos (ano)`,
		`CREATE TABLE IF NOT EXISTS passagens (
			id                    CHAR(40)    PRIMARY KEY,
			processo_id           VARCHAR(80) NOT NULL,
			pcdp                  VARCHAR(80) NOT NULL,
			meio_transporte       VARCHAR(120),
			ida_origem_pais       VARCHAR(120),
			ida_origem_uf         VARCHAR(40),
			ida_origem_cidade     VARCHAR(255),
			ida_destino_pais      VARCHAR(120),
			ida_destino_uf        VARCHAR(40),
			ida_destino_cidade    VARCHAR(255),
			volta_origem_pais     VARCHAR(120),
			volta_origem_uf       VARCHAR(40),
			volta_origem_cidade   VARCHAR(255),
			volta_destino_pais    VARCHAR(120),
			volta_destino_uf      VARCHAR(40),
			volta_destino_cidade  VARCHAR(255),
			valor_passagem_cents  BIGINT,
			taxa_servico_cents    BIGINT,
			emissao_data          DATE,
			emissao_hora          VARCHAR(40),
			ano                   SMALLINT    NOT NULL,
			imported_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_passagens_processo ON passagens (processo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_passagens_emissao_data ON passagens (emissao_data)`,
		`CREATE INDEX IF NOT EXISTS idx_passagens_ano ON passagens (ano)`,
		`CREATE TABLE IF NOT EXISTS pagamentos (
			id                    CHAR(40)    PRIMARY KEY,
			processo_id           VARCHAR(80) NOT NULL,
			pcdp                  VARCHAR(80) NOT NULL,
			orgao_superior_codigo VARCHAR(30),
			orgao_superior_nome   VARCHAR(255),
			orgao_pagador_codigo  VARCHAR(30),
			orgao_pagador_nome    VARCHAR(255),
			ug_pagadora_codigo    VARCHAR(40),
			ug_pagadora_nome      VARCHAR(255),
			tipo_pagamento        VARCHAR(120),
			valor_cents           BIGINT,
			ano                   SMALLINT    NOT NULL,
			imported_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_pagamentos_processo ON pagamentos (processo_id)`,
		`CREATE INDEX IF NOT EXISTS idx_pagamentos_tipo ON pagamentos (tipo_pagamento)`,
		`CREATE INDEX IF NOT EXISTS idx_pagamentos_ano ON pagamentos (ano)`,
		`CREATE TABLE IF NOT EXISTS imports (
			ano            SMALLINT    PRIMARY KEY,
			zip_sha256     CHAR(64)    NOT NULL,
			status         VARCHAR(20) NOT NULL CHECK (status IN ('downloading','extracted','importing','done','error')),
			last_step      VARCHAR(20) CHECK (last_step IN ('download','extract','viagem','trecho','passagem','pagamento')),
			rows_viagem    BIGINT DEFAULT 0,
			rows_trecho    BIGINT DEFAULT 0,
			rows_passagem  BIGINT DEFAULT 0,
			rows_pagamento BIGINT DEFAULT 0,
			started_at     TIMESTAMPTZ,
			finished_at    TIMESTAMPTZ,
			error_msg      TEXT
		)`,
		`ALTER TABLE IF EXISTS trechos ALTER COLUMN origem_uf TYPE VARCHAR(40)`,
		`ALTER TABLE IF EXISTS trechos ALTER COLUMN destino_uf TYPE VARCHAR(40)`,
		`ALTER TABLE IF EXISTS passagens ALTER COLUMN ida_origem_uf TYPE VARCHAR(40)`,
		`ALTER TABLE IF EXISTS passagens ALTER COLUMN ida_destino_uf TYPE VARCHAR(40)`,
		`ALTER TABLE IF EXISTS passagens ALTER COLUMN volta_origem_uf TYPE VARCHAR(40)`,
		`ALTER TABLE IF EXISTS passagens ALTER COLUMN volta_destino_uf TYPE VARCHAR(40)`,
		`ALTER TABLE IF EXISTS passagens ALTER COLUMN emissao_hora TYPE VARCHAR(40)`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
