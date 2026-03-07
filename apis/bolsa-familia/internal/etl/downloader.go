// Radar do Povo ETL - https://radardopovo.com
package etl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"log/slog"

	"github.com/radardopovo/bolsa-familia-etl/internal/db"
)

func (o *Orchestrator) ensurePeriodZIP(ctx context.Context, period int, dstPath string, st *db.ImportState) (string, error) {
	restartDownload := st != nil && st.Status == "downloading"
	if !restartDownload {
		if localSHA, ok := fileSHA256IfExists(dstPath); ok {
			if st != nil && st.ZipSHA256 != "" && localSHA == st.ZipSHA256 && !o.cfg.Force {
				o.log.Info("zip_cache_hit",
					slog.Int("periodo", period),
					slog.String("zip", dstPath),
					slog.String("sha256", localSHA),
				)
				return localSHA, nil
			}
			if st == nil || st.ZipSHA256 == "" {
				o.log.Info("zip_local_reuse_without_state",
					slog.Int("periodo", period),
					slog.String("zip", dstPath),
					slog.String("sha256", localSHA),
				)
				return localSHA, nil
			}
		}
	}

	url := fmt.Sprintf("https://portaldatransparencia.gov.br/download-de-dados/novo-bolsa-familia/%d", period)
	resp, err := o.client.Get(ctx, url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download falhou: status %d", resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return "", err
	}
	tmpPath := dstPath + ".tmp"
	tmp, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	head := make([]byte, 4)
	if _, err := io.ReadFull(resp.Body, head); err != nil {
		return "", fmt.Errorf("nao foi possivel ler magic bytes do ZIP: %w", err)
	}
	if string(head) != "PK\x03\x04" {
		return "", fmt.Errorf("resposta nao e um ZIP valido para o periodo %d", period)
	}

	hasher := sha256.New()
	mw := io.MultiWriter(tmp, hasher)
	if _, err := mw.Write(head); err != nil {
		return "", err
	}
	n, err := io.Copy(mw, resp.Body)
	if err != nil {
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmpPath, dstPath); err != nil {
		return "", err
	}

	sha := hex.EncodeToString(hasher.Sum(nil))
	o.log.Info("zip_downloaded",
		slog.Int("periodo", period),
		slog.String("zip", dstPath),
		slog.Int64("bytes", n+4),
		slog.String("sha256", sha),
		slog.Time("at", time.Now().UTC()),
	)
	return sha, nil
}

func fileSHA256IfExists(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", false
	}
	return hex.EncodeToString(h.Sum(nil)), true
}
