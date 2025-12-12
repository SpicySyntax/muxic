package main

import (
	"os"
	"testing"
)

func TestConfig_SaveAndLoad(t *testing.T) {
	// Setup: Remove existing config file if present (backup could be handled, but for test simplicity we just clean up)
	// In a real scenario, we might want to use a temp dir or mock the filesystem, but LoadConfig hardcodes "muxic_config.json"
	// For this test, we will rename the real config if it exists, and restore it after.

	if _, err := os.Stat(configFileName); err == nil {
		os.Rename(configFileName, configFileName+".bak")
		defer os.Rename(configFileName+".bak", configFileName)
	} else {
		// If it didn't exist, ensure we clean up the test file
		defer os.Remove(configFileName)
	}

	deviceID := "{0.0.0.00000000}.{some-guid}"
	cfg := Config{
		DefaultDevice: &deviceID,
	}

	// Test Save
	err := cfg.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}

	// Test Load
	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.DefaultDevice == nil {
		t.Fatal("Loaded DefaultDevice is nil")
	}
	if *loadedCfg.DefaultDevice != deviceID {
		t.Errorf("Expected DefaultDevice %v, got %v", deviceID, *loadedCfg.DefaultDevice)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	// Ensure no config file exists
	if _, err := os.Stat(configFileName); err == nil {
		os.Rename(configFileName, configFileName+".bak")
		defer os.Rename(configFileName+".bak", configFileName)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed when no file exists: %v", err)
	}
	if cfg.DefaultDevice != nil {
		t.Error("Expected nil DefaultDevice when no config file exists")
	}
}
