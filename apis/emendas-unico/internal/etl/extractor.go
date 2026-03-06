// Radar do Povo ETL - https://radardopovo.com
package etl

import "github.com/radardopovo/emendas-etl/internal/zipx"

func detectExtractedCSVs(extractPath string) (zipx.EmendasCSVFiles, error) {
	return zipx.DetectInDirectory(extractPath)
}

func extractCSVsFromZIP(zipPath, extractPath string) (zipx.EmendasCSVFiles, error) {
	return zipx.ExtractCSVFiles(zipPath, extractPath)
}
