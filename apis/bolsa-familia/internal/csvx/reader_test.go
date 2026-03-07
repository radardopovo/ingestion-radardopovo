// Radar do Povo ETL - https://radardopovo.com
package csvx

import "testing"

func TestNormalizeHeader(t *testing.T) {
	in := "\uFEFF  Número   da Proposta (PCDP)  "
	got := NormalizeHeader(in)
	want := "numero da proposta (pcdp)"
	if got != want {
		t.Fatalf("NormalizeHeader() = %q want %q", got, want)
	}
}
