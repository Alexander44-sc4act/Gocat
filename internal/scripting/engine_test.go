package scripting

import "testing"

func TestLuaEngineExecuteScriptPropagatesMainError(t *testing.T) {
	eng := NewLuaEngine(nil)
	defer eng.Close()

	if err := eng.LoadString(`function main() error("boom") end`); err != nil {
		t.Fatalf("LoadString: %v", err)
	}
	if err := eng.ExecuteScript("failing"); err == nil {
		t.Fatal("ExecuteScript(failing main) = nil, want error")
	}
}

func TestLuaEngineExecuteScriptNoMainSucceeds(t *testing.T) {
	eng := NewLuaEngine(nil)
	defer eng.Close()

	if err := eng.LoadString(`x = 1 + 1`); err != nil {
		t.Fatalf("LoadString: %v", err)
	}
	if err := eng.ExecuteScript("no-main"); err != nil {
		t.Fatalf("ExecuteScript(no main) = %v, want nil", err)
	}
}
