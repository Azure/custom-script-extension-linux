package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseCmd(t *testing.T) {
	const errString = "shlex: failed to parse command: "

	cases := []struct {
		in        string
		out       []string
		expectErr bool
	}{
		{"one", []string{"one"}, false},                            // one arg
		{"one two", []string{"one", "two"}, false},                 // two args
		{"  one", []string{"one"}, false},                          // trim left space
		{`\ \ one`, []string{"  one"}, false},                      // escaped space preserved
		{`one \ `, []string{"one", " "}, false},                    // first arg is one space char
		{`one\ two \ three`, []string{"one two", " three"}, false}, // prg name contains space + first arg starts with space

		{`one "two three"`, []string{"one", "two three"}, false},                    // simple double-quote
		{`one "two three 'four'"`, []string{"one", "two three 'four'"}, false},      // single quote inside double quote
		{`one 'two' "three' four'"`, []string{"one", "two", "three' four'"}, false}, // mixed quotes
		{`one \""two`, nil, true},                                                   // missing quote
		{`one \""two\""`, []string{"one", `"two"`}, false},                          // escaped double-quotes preserved
		{`one \''two\''`, nil, true},                                                // misuse of single quote
		{`/usr/bin/env python -c 'import sys; sys.exit(1)'`, []string{
			"/usr/bin/env", "python", "-c", "import sys; sys.exit(1)"}, false}, // longer command

		// super long command making a POST request
		{`curl 'https://sourcegraph.com/.api/repos' -H 'cookie: a=b; c=d' -H 'pragma: no-cache' -H 'accept: */*' -H 'cache-control: no-cache' -H 'authority: sourcegraph.com' -H 'referer: https://github.com/flynn-archive/go-shlex/blob/master/shlex_test.go' --data-binary '{"Op":{"New":{"URI":"github.com/flynn-archive/go-shlex","CloneURL":"https://github.com/flynn-archive/go-shlex","DefaultBranch":"master","Mirror":true}}}' --compressed`, []string{
			"curl", "https://sourcegraph.com/.api/repos",
			"-H", "cookie: a=b; c=d", "-H", "pragma: no-cache",
			"-H", "accept: */*", "-H", "cache-control: no-cache",
			"-H", "authority: sourcegraph.com",
			"-H", "referer: https://github.com/flynn-archive/go-shlex/blob/master/shlex_test.go",
			"--data-binary", `{"Op":{"New":{"URI":"github.com/flynn-archive/go-shlex","CloneURL":"https://github.com/flynn-archive/go-shlex","DefaultBranch":"master","Mirror":true}}}`,
			"--compressed"}, false},
	}

	for i, c := range cases {
		v, err := parseCmd(c.in)
		if err != nil {
			if c.expectErr {
				require.Contains(t, err.Error(), errString, c.in)
			} else {
				require.Nil(t, err, "unexpected failure. case=%d cmd=%s", i, c.in)
			}
		} else {
			if c.expectErr {
				require.Fail(t, "expected error, got none", "case=%d cmd=%s", i, c.in)
			}
			require.Equal(t, c.out, v, "wrong values. case=%d cmd=%s", i, c.in)
		}
	}
}
