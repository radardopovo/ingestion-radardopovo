// Radar do Povo ETL - https://radardopovo.com
package etl

import "github.com/radardopovo/viagens-etl/internal/zipx"

func detectExtractedCSVs(extractPath string) (zipx.CSVFiles, error) {
	return zipx.DetectInDirectory(extractPath)
}

func extractCSVsFromZIP(zipPath, extractPath string) (zipx.CSVFiles, error) {
	return zipx.ExtractCSVFiles(zipPath, extractPath)
}
