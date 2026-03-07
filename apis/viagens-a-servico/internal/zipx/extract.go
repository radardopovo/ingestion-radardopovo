// Radar do Povo ETL - https://radardopovo.com
package zipx

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CSVFiles struct {
	Viagem    string
	Trecho    string
	Passagem  string
	Pagamento string
}

func ValidateZipMagic(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := make([]byte, 4)
	if _, err := io.ReadFull(f, buf); err != nil {
		return err
	}
	if string(buf) != "PK\x03\x04" {
		return fmt.Errorf("arquivo nao e ZIP valido: %s", path)
	}
	return nil
}

func ExtractCSVFiles(zipPath, outDir string) (CSVFiles, error) {
	if err := ValidateZipMagic(zipPath); err != nil {
		return CSVFiles{}, err
	}
	if err := os.RemoveAll(outDir); err != nil {
		return CSVFiles{}, err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return CSVFiles{}, err
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return CSVFiles{}, err
	}
	defer r.Close()

	files := CSVFiles{}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		base := filepath.Base(f.Name)
		if !strings.HasSuffix(strings.ToLower(base), ".csv") {
			continue
		}
		dstPath := filepath.Join(outDir, base)
		if err := extractOne(f, dstPath); err != nil {
			return CSVFiles{}, err
		}
		switch detectRole(base) {
		case "viagem":
			files.Viagem = dstPath
		case "trecho":
			files.Trecho = dstPath
		case "passagem":
			files.Passagem = dstPath
		case "pagamento":
			files.Pagamento = dstPath
		}
	}

	if err := validateDetected(files); err != nil {
		return CSVFiles{}, err
	}
	return files, nil
}

func DetectInDirectory(dir string) (CSVFiles, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return CSVFiles{}, err
	}
	files := CSVFiles{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".csv") {
			continue
		}
		full := filepath.Join(dir, name)
		switch detectRole(name) {
		case "viagem":
			files.Viagem = full
		case "trecho":
			files.Trecho = full
		case "passagem":
			files.Passagem = full
		case "pagamento":
			files.Pagamento = full
		}
	}
	if err := validateDetected(files); err != nil {
		return CSVFiles{}, err
	}
	return files, nil
}

func detectRole(fileName string) string {
	n := strings.ToLower(fileName)
	switch {
	case strings.Contains(n, "trecho"):
		return "trecho"
	case strings.Contains(n, "passagem"):
		return "passagem"
	case strings.Contains(n, "pagamento"):
		return "pagamento"
	case strings.Contains(n, "viagem"):
		return "viagem"
	default:
		return ""
	}
}

func validateDetected(files CSVFiles) error {
	missing := make([]string, 0, 4)
	if files.Viagem == "" {
		missing = append(missing, "Viagem")
	}
	if files.Trecho == "" {
		missing = append(missing, "Trecho")
	}
	if files.Passagem == "" {
		missing = append(missing, "Passagem")
	}
	if files.Pagamento == "" {
		missing = append(missing, "Pagamento")
	}
	if len(missing) > 0 {
		return errors.New("CSVs obrigatorios nao encontrados: " + strings.Join(missing, ", "))
	}
	return nil
}

func extractOne(file *zip.File, dstPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return nil
}
