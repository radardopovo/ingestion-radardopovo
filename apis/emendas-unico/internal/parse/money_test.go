// Radar do Povo ETL - https://radardopovo.com
package parse

import "testing"

func TestMoneyToCents(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"1.234,56", 123456, true},
		{"1234,56", 123456, true},
		{"1234.56", 123456, true},
		{"0,00", 0, true},
		{"", 0, false},
		{"Sem informacao", 0, false},
		{"Sem informação", 0, false},
		{"S/I", 0, false},
		{"-500,00", -50000, true},
	}

	for _, tc := range cases {
		got, ok := MoneyToCents(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("MoneyToCents(%q) = (%d,%v), want (%d,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}
