// Radar do Povo ETL - https://radardopovo.com
package parse

import (
	"testing"
	"time"
)

func TestDateBR(t *testing.T) {
	got, ok := DateBR("05/03/2026")
	if !ok {
		t.Fatal("expected valid date")
	}
	want := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("DateBR mismatch: got %v want %v", got, want)
	}

	if _, ok := DateBR(""); ok {
		t.Fatal("expected empty date to be invalid")
	}
}
