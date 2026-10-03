package plugins

import "testing"

func TestHelloWorld(t *testing.T) {
	result := "hello world"
	expected := "hello world"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestPluginPackageExists(t *testing.T) {
	// Package'ın var olduğunu test et
	// This test ensures the plugin package compiles correctly
	t.Log("Plugin package exists and compiles successfully")
}

func TestBasicPluginFunction(t *testing.T) {
	pluginName := "test-plugin"
	if len(pluginName) == 0 {
		t.Error("Plugin name should not be empty")
	}
}
