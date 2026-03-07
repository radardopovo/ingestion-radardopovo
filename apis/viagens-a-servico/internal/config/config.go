// Radar do Povo ETL - https://radardopovo.com
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DBHost               string
	DBPort               int
	DBUser               string
	DBPass               string
	DBName               string
	DBSSLCA              string
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetimeMin int

	DataDir        string
	BatchSize      int
	ChunkSize      int
	WriterWorkers  int
	HTTPTimeoutSec int
	HTTPMaxRetries int

	FromYear int
	ToYear   int
	Year     int

	Force        bool
	OnlyDownload bool
	OnlyImport   bool
	InsertMode   bool
	Verbose      bool
	LogJSON      bool
	DryRun       bool
}

func Load(args []string, now time.Time) (Config, error) {
	currentYear := now.UTC().Year()
	cfg := Config{
		DBHost:               envString("DB_HOST", ""),
		DBPort:               envInt("DB_PORT", 5432),
		DBUser:               envString("DB_USER", ""),
		DBPass:               envString("DB_PASS", ""),
		DBName:               envString("DB_NAME", ""),
		DBSSLCA:              envString("DB_SSL_CA", ""),
		DBMaxOpenConns:       envInt("DB_MAX_OPEN_CONNS", 30),
		DBMaxIdleConns:       envInt("DB_MAX_IDLE_CONNS", 30),
		DBConnMaxLifetimeMin: envInt("DB_CONN_MAX_LIFETIME_MIN", 10),
		DataDir:              envString("DATA_DIR", "./data"),
		BatchSize:            envInt("BATCH_SIZE", 1000),
		ChunkSize:            envInt("CHUNK_SIZE", 5000),
		WriterWorkers:        envInt("WRITER_WORKERS", 2),
		HTTPTimeoutSec:       envInt("HTTP_TIMEOUT_SEC", 120),
		HTTPMaxRetries:       envInt("HTTP_MAX_RETRIES", 5),
		FromYear:             2011,
		ToYear:               currentYear,
	}

	fs := flag.NewFlagSet("importer", flag.ContinueOnError)
	fs.IntVar(&cfg.FromYear, "from", cfg.FromYear, "Ano inicial (inclusive)")
	fs.IntVar(&cfg.ToYear, "to", cfg.ToYear, "Ano final (inclusive)")
	fs.IntVar(&cfg.Year, "year", 0, "Importar apenas um ano especifico")
	fs.BoolVar(&cfg.Force, "force", false, "Reimportar mesmo status done")
	fs.BoolVar(&cfg.OnlyDownload, "only-download", false, "Baixar e extrair sem importar")
	fs.BoolVar(&cfg.OnlyImport, "only-import", false, "Nao baixar, apenas importar ZIP/CSV locais")
	fs.BoolVar(&cfg.InsertMode, "insert-mode", false, "Usar INSERT multi-row em vez de COPY")
	fs.IntVar(&cfg.BatchSize, "batch-size", cfg.BatchSize, "Linhas por batch")
	fs.IntVar(&cfg.ChunkSize, "chunk-size", cfg.ChunkSize, "Linhas por commit")
	fs.IntVar(&cfg.WriterWorkers, "writers", cfg.WriterWorkers, "Numero de workers de escrita por CSV")
	fs.StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "Diretorio raiz dos dados")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "Log detalhado")
	fs.BoolVar(&cfg.LogJSON, "log-json", envBool("LOG_JSON", false), "Emitir logs em JSON")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "Parseia sem gravar no banco")
	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	if cfg.Year != 0 {
		cfg.FromYear = cfg.Year
		cfg.ToYear = cfg.Year
	}

	if err := cfg.validate(currentYear); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Years() []int {
	years := make([]int, 0, c.ToYear-c.FromYear+1)
	for y := c.FromYear; y <= c.ToYear; y++ {
		years = append(years, y)
	}
	return years
}

func (c Config) DBConnMaxLifetime() time.Duration {
	return time.Duration(c.DBConnMaxLifetimeMin) * time.Minute
}

func (c Config) ModeLabel() string {
	if c.InsertMode {
		return "INSERT"
	}
	return "COPY"
}

func (c Config) validate(currentYear int) error {
	if c.OnlyDownload && c.OnlyImport {
		return fmt.Errorf("--only-download e --only-import nao podem ser usados juntos")
	}
	if c.FromYear < 2011 || c.ToYear < 2011 {
		return fmt.Errorf("anos devem ser >= 2011")
	}
	if c.FromYear > c.ToYear {
		return fmt.Errorf("--from nao pode ser maior que --to")
	}
	if c.ToYear > currentYear {
		return fmt.Errorf("--to nao pode ser maior que o ano corrente (%d)", currentYear)
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("--batch-size deve ser > 0")
	}
	if c.ChunkSize <= 0 {
		return fmt.Errorf("--chunk-size deve ser > 0")
	}
	if c.WriterWorkers <= 0 {
		return fmt.Errorf("--writers deve ser > 0")
	}
	if c.HTTPTimeoutSec <= 0 {
		return fmt.Errorf("HTTP_TIMEOUT_SEC deve ser > 0")
	}
	if c.HTTPMaxRetries < 0 {
		return fmt.Errorf("HTTP_MAX_RETRIES deve ser >= 0")
	}
	if strings.TrimSpace(c.DataDir) == "" {
		return fmt.Errorf("--data-dir nao pode ser vazio")
	}
	if c.DryRun {
		return nil
	}
	if strings.TrimSpace(c.DBHost) == "" ||
		strings.TrimSpace(c.DBUser) == "" ||
		strings.TrimSpace(c.DBPass) == "" ||
		strings.TrimSpace(c.DBName) == "" ||
		strings.TrimSpace(c.DBSSLCA) == "" {
		return fmt.Errorf("variaveis DB_HOST, DB_USER, DB_PASS, DB_NAME e DB_SSL_CA sao obrigatorias")
	}
	if c.DBPort <= 0 {
		return fmt.Errorf("DB_PORT deve ser > 0")
	}
	if c.DBMaxOpenConns <= 0 || c.DBMaxIdleConns <= 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS e DB_MAX_IDLE_CONNS devem ser > 0")
	}
	if c.DBConnMaxLifetimeMin <= 0 {
		return fmt.Errorf("DB_CONN_MAX_LIFETIME_MIN deve ser > 0")
	}
	return nil
}

func envString(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func envInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	iv, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return iv
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}
