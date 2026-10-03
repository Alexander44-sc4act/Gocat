package session

import "testing"

func TestShellQuote(t *testing.T) {
	cases := []struct{ in, want string }{
		{"echo hi", "'echo hi'"},
		{`echo "hi" $HOME`, `'echo "hi" $HOME'`},
		{"it's", `'it'"'"'s'`},
		{"", "''"},
	}
	for _, tc := range cases {
		if got := ShellQuote(tc.in); got != tc.want {
			t.Errorf("ShellQuote(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
