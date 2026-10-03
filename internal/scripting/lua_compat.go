package scripting

import (
	"fmt"

	"github.com/realibrahimsql/Gocat/internal/scripting/modules"
	lua "github.com/yuin/gopher-lua"
)

// LuaEngine is a compatibility wrapper for the old API
// DEPRECATED: Use Engine instead
type LuaEngine struct {
	*Engine
}

// EngineConfig is a compatibility alias
// DEPRECATED: Use Config instead
type EngineConfig = Config

// NewLuaEngine creates a new Lua engine (compatibility function)
// DEPRECATED: Use NewEngine instead
func NewLuaEngine(config *EngineConfig) *LuaEngine {
	return &LuaEngine{
		Engine: NewEngine(config),
	}
}

// Compatibility exports for backward compatibility
var (
	// Export module registration functions for packages that might use them directly
	RegisterNetworkModule = modules.RegisterNetworkModule
	RegisterHTTPModule    = modules.RegisterHTTPModule
	RegisterCryptoModule  = modules.RegisterCryptoModule
	RegisterSystemModule  = modules.RegisterSystemModule
	RegisterFileModule    = modules.RegisterFileModule
	RegisterTimeModule    = modules.RegisterTimeModule
	RegisterUIModule      = modules.RegisterUIModule
	RegisterJSONModule    = modules.RegisterJSONModule
)

// GetLuaState returns the underlying Lua state
// This is for advanced users who need direct access
func (e *LuaEngine) GetLuaState() *lua.LState {
	if e.Engine != nil {
		return e.Engine.L
	}
	return nil
}

// ExecuteScript executes a loaded script's main function
// DEPRECATED: Use ExecuteFunction("main") instead
func (e *LuaEngine) ExecuteScript(scriptName string) error {
	if e.Engine == nil || e.Engine.L == nil {
		return fmt.Errorf("lua engine is closed")
	}
	// Scripts without a main function already ran their top-level code
	// during load; there is nothing more to execute.
	if fn := e.Engine.L.GetGlobal("main"); fn.Type() != lua.LTFunction {
		return nil
	}
	_, err := e.ExecuteFunction("main")
	return err
}
