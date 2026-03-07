// Radar do Povo ETL - https://radardopovo.com
package db

import "testing"

func TestSafeInsertBatchSize(t *testing.T) {
	if got := safeInsertBatchSize(5000, 24); got != 2730 {
		t.Fatalf("safeInsertBatchSize(5000,24) = %d want 2730", got)
	}
	if got := safeInsertBatchSize(1000, 10); got != 1000 {
		t.Fatalf("safeInsertBatchSize(1000,10) = %d want 1000", got)
	}
	if got := safeInsertBatchSize(0, 10); got != 1 {
		t.Fatalf("safeInsertBatchSize(0,10) = %d want 1", got)
	}
}
