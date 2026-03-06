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

type EmendasCSVFiles struct {
	Emendas      string
	PorFavorecido string
	Convenios    string
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

func ExtractCSVFiles(zipPath, outDir string) (EmendasCSVFiles, error) {
	if err := ValidateZipMagic(zipPath); err != nil {
		return EmendasCSVFiles{}, err
	}
	if err := os.RemoveAll(outDir); err != nil {
		return EmendasCSVFiles{}, err
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return EmendasCSVFiles{}, err
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return EmendasCSVFiles{}, err
	}
	defer r.Close()

	files := EmendasCSVFiles{}
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
			return EmendasCSVFiles{}, err
		}

		switch detectRole(base) {
		case "convenios":
			files.Convenios = dstPath
		case "favorecido":
			files.PorFavorecido = dstPath
		case "emendas":
			files.Emendas = dstPath
		}
	}

	if err := validateDetected(files); err != nil {
		return EmendasCSVFiles{}, err
	}
	return files, nil
}

func DetectInDirectory(dir string) (EmendasCSVFiles, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return EmendasCSVFiles{}, err
	}

	files := EmendasCSVFiles{}
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
		case "convenios":
			files.Convenios = full
		case "favorecido":
			files.PorFavorecido = full
		case "emendas":
			files.Emendas = full
		}
	}

	if err := validateDetected(files); err != nil {
		return EmendasCSVFiles{}, err
	}
	return files, nil
}

func detectRole(fileName string) string {
	n := strings.ToLower(fileName)
	n = strings.ReplaceAll(n, "_", "")
	n = strings.ReplaceAll(n, "-", "")
	switch {
	case strings.Contains(n, "convenio"):
		return "convenios"
	case strings.Contains(n, "porfavorecido"):
		return "favorecido"
	case strings.Contains(n, "emendasparlamentares"):
		return "emendas"
	default:
		return ""
	}
}

func validateDetected(files EmendasCSVFiles) error {
	missing := make([]string, 0, 3)
	if files.Emendas == "" {
		missing = append(missing, "EmendasParlamentares.csv")
	}
	if files.PorFavorecido == "" {
		missing = append(missing, "EmendasParlamentares_PorFavorecido.csv")
	}
	if files.Convenios == "" {
		missing = append(missing, "EmendasParlamentares_Convenios.csv")
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
