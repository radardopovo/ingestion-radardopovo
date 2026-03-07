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
	CPGF string
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
		if detectRole(base) == "cpgf" {
			files.CPGF = dstPath
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
		if detectRole(name) == "cpgf" {
			files.CPGF = full
		}
	}
	if err := validateDetected(files); err != nil {
		return CSVFiles{}, err
	}
	return files, nil
}

func detectRole(fileName string) string {
	n := strings.ToLower(fileName)
	if strings.Contains(n, "cpgf") {
		return "cpgf"
	}
	return ""
}

func validateDetected(files CSVFiles) error {
	if files.CPGF == "" {
		return errors.New("CSV obrigatorio nao encontrado: CPGF")
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
