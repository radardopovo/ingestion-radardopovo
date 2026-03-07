// Radar do Povo ETL - https://radardopovo.com
package parse

import (
	"crypto/sha1"
	"fmt"
)

func MakeID(parts ...string) string {
	h := sha1.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
