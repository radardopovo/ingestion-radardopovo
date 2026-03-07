// Radar do Povo ETL - https://radardopovo.com
package parse

import (
	"strings"
	"time"
)

func DateBR(s string) (time.Time, bool) {
	v := strings.TrimSpace(s)
	if v == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("02/01/2006", v)
	if err != nil {
		return time.Time{}, false
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), true
}
