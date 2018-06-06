package main

import (
	"testing"

	"github.com/spf13/viper"
)

func TestNrParsing(t *testing.T) {
	viper.SetDefault("rev_pattern", revPattern)
	viper.SetDefault("nr_pattern", nrPattern)

	cases := []struct {
		in, nr, rev string
	}{
		{"P1234-M1230-AA_dsfsdaf.pdf", "P1234-M1230", "AA"},
		{"p1234-1235_AA-dsfsdaf.pdf", "p1234-1235", "AA"},
		{"p3234-M123-AA.pdf", "p3234-M123", "AA"},
		{"P121-C223.cd", "P121-C223", ""},
		{"P121-C223", "P121-C223", ""},
		{"P123-325_AA", "P123-325", "AA"},
	}

	for _, c := range cases {
		nr, rev, err := parseDocNr(c.in)
		if err != nil {
			t.Errorf("%s", err)
		}
		if nr != c.nr {
			t.Errorf("getNr(%q) == %q, want %q", c.in, nr, c.nr)
		}
		if rev != c.rev {
			t.Errorf("getRev(%q) == %q, want %q", c.in, rev, c.rev)
		}
	}
}
