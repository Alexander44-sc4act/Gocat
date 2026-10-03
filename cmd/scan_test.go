package cmd

import (
	"reflect"
	"testing"
)

// TestParsePortRange guards the port-parsing edge cases (bounds, empty items,
// reversed ranges, dedup) that a typo in a scan spec can otherwise turn into a
// silent no-op or an out-of-range dial.
func TestParsePortRange(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{name: "single port", input: "80", want: []int{80}},
		{name: "comma list", input: "22,80,443", want: []int{22, 80, 443}},
		{name: "comma list with spaces", input: "22 , 80 , 443", want: []int{22, 80, 443}},
		{name: "range", input: "80-82", want: []int{80, 81, 82}},
		{name: "mixed range and port", input: "80-82,443", want: []int{80, 81, 82, 443}},
		{name: "duplicates deduped", input: "80,80,80-81", want: []int{80, 81}},
		{name: "full range", input: "1-1000", want: portSeq(1, 1000)},
		{name: "empty", input: "", wantErr: true},
		{name: "whitespace only", input: "   ", wantErr: true},
		{name: "zero port", input: "0", wantErr: true},
		{name: "port too large", input: "65536", wantErr: true},
		{name: "negative port", input: "-1", wantErr: true},
		{name: "empty list element", input: "22,,80", wantErr: true},
		{name: "reversed range", input: "10-5", wantErr: true},
		{name: "range out of bounds", input: "65530-65540", wantErr: true},
		{name: "non numeric", input: "http", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePortRange(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parsePortRange(%q) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parsePortRange(%q) unexpected error: %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parsePortRange(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func portSeq(start, end int) []int {
	out := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		out = append(out, i)
	}
	return out
}
