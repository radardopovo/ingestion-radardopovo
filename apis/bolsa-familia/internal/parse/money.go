// Radar do Povo ETL - https://radardopovo.com
package parse

import (
	"math"
	"strconv"
	"strings"
)

func MoneyToCents(s string) (int64, bool) {
	v := strings.TrimSpace(strings.ReplaceAll(s, "\u00a0", " "))
	if v == "" {
		return 0, false
	}
	if strings.EqualFold(v, "Sem informa\u00e7\u00e3o") || strings.EqualFold(v, "Sem informacao") || strings.EqualFold(v, "S/I") {
		return 0, false
	}
	v = strings.ReplaceAll(v, "R$", "")
	v = strings.ReplaceAll(v, " ", "")
	if strings.Contains(v, ",") {
		v = strings.ReplaceAll(v, ".", "")
	}
	v = strings.ReplaceAll(v, ",", ".")
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false
	}
	return int64(math.Round(f * 100)), true
}
