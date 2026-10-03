package modules

import (
	"strings"
	"testing"
)

func TestSearchFindsByNameDescriptionCategory(t *testing.T) {
	if got := DefaultRegistry.Search("linpeas"); len(got) == 0 {
		t.Fatal("Search(linpeas) returned nothing")
	}
	if got := DefaultRegistry.Search("credential"); len(got) == 0 {
		t.Fatal("Search(credential) returned nothing")
	}
	if got := DefaultRegistry.Search("active directory"); len(got) == 0 {
		t.Fatal("Search(active directory) returned nothing")
	}
	if got := DefaultRegistry.Search("definitely-not-a-module-xyz"); len(got) != 0 {
		t.Fatalf("Search(unknown) = %d, want 0", len(got))
	}
}

func TestDescribeKnownAndUnknown(t *testing.T) {
	out := DefaultRegistry.Describe("linpeas")
	if !strings.Contains(out, "linpeas") || !strings.Contains(out, "run linpeas") {
		t.Fatalf("Describe(linpeas) missing details:\n%s", out)
	}
	out = DefaultRegistry.Describe("no-such-module")
	if !strings.Contains(out, "not found") {
		t.Fatalf("Describe(unknown) should report not found:\n%s", out)
	}
}

func TestWindowsModulesRegistered(t *testing.T) {
	for _, name := range []string{"sigmapotato", "adpeas", "seatbelt", "meterpreter"} {
		if mod := DefaultRegistry.Get(name); mod == nil {
			t.Errorf("module %s not registered", name)
		} else if !mod.Enabled {
			t.Errorf("module %s disabled", name)
		}
	}
}

func TestEnumerationModulesRegistered(t *testing.T) {
	for _, name := range []string{"enumerate", "report", "escalate", "implant", "tamper"} {
		if mod := DefaultRegistry.Get(name); mod == nil {
			t.Errorf("module %s not registered", name)
		} else if !mod.Enabled {
			t.Errorf("module %s disabled", name)
		}
	}
}
