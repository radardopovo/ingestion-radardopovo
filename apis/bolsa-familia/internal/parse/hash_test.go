// Radar do Povo ETL - https://radardopovo.com
package parse

import "testing"

func TestMakeIDDeterministic(t *testing.T) {
	a := MakeID("x", "y", "z")
	b := MakeID("x", "y", "z")
	c := MakeID("x", "y", "w")

	if a != b {
		t.Fatal("same parts should produce same hash")
	}
	if a == c {
		t.Fatal("different parts should produce different hash")
	}
	if len(a) != 40 {
		t.Fatalf("sha1 hex should be 40 chars, got %d", len(a))
	}
}
