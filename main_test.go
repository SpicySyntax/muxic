package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/moutend/go-wca/pkg/wca"
)

func TestGetTrackPath(t *testing.T) {
	trackName := "test_track"
	expected := filepath.Join(TracksDir, trackName+".wav")
	result := getTrackPath(trackName)

	if result != expected {
		t.Errorf("Expected path %s, got %s", expected, result)
	}
}

func TestEnsureTracksDir(t *testing.T) {
	// Setup: Remove tracks dir if it exists
	if _, err := os.Stat(TracksDir); err == nil {
		// Just in case implementation changes to error on existing file
	}
	// We won't delete the real tracks dir to avoid data loss,
	// but we can check if it ensures it exists.
	// Actually, `ensureTracksDir` creates it if missing.

	// Let's rely on the fact that it should succeed whether it exists or not.
	err := ensureTracksDir()
	if err != nil {
		t.Errorf("ensureTracksDir failed: %v", err)
	}

	// Verify it exists
	info, err := os.Stat(TracksDir)
	if err != nil {
		t.Errorf("TracksDir does not exist after ensureTracksDir")
	}
	if !info.IsDir() {
		t.Errorf("TracksDir is not a directory")
	}
}

func TestSaveWavFile(t *testing.T) {
	// Create a temp file path
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_output.wav")
	defer os.Remove(tmpFile)

	// dummy audio data
	audioData := []byte{0x01, 0x02, 0x03, 0x04}

	// dummy wfx
	wfx := &wca.WAVEFORMATEX{
		WFormatTag:      1, // PCM
		NChannels:       2,
		NSamplesPerSec:  44100,
		NAvgBytesPerSec: 176400,
		NBlockAlign:     4,
		WBitsPerSample:  16,
		CbSize:          0,
	}

	err := saveWavFile(tmpFile, audioData, wfx)
	if err != nil {
		t.Fatalf("saveWavFile failed: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read created wav file: %v", err)
	}

	// Header is 44 bytes, data is 4 bytes = 48 bytes
	if len(content) != 44+4 {
		t.Errorf("Expected file size 48, got %d", len(content))
	}

	// Verify "RIFF" header
	if string(content[0:4]) != "RIFF" {
		t.Errorf("Invalid RIFF header")
	}

	// Verify "WAVE" format
	if string(content[8:12]) != "WAVE" {
		t.Errorf("Invalid WAVE header")
	}
}

func TestCalculateAmplitude(t *testing.T) {
	// Test 16-bit silence
	silence16 := []byte{0, 0, 0, 0}
	amp := calculateAmplitude(silence16, 16)
	if amp != 0 {
		t.Errorf("Expected 0 amplitude for silence, got %f", amp)
	}

	// Test 16-bit max amplitude
	// 32767 = 0x7FFF -> 0xFF 0x7F (Little Endian)
	max16 := []byte{0xFF, 0x7F}
	amp = calculateAmplitude(max16, 16)
	// 32767/32768 ~= 0.999969
	if amp < 0.99 {
		t.Errorf("Expected ~1.0 amplitude for max int16, got %f", amp)
	}

	// Test 32-bit float max amplitude
	// 1.0 = 0x3F800000 -> 00 00 80 3F
	max32 := []byte{0x00, 0x00, 0x80, 0x3F}
	amp = calculateAmplitude(max32, 32)
	if amp != 1.0 {
		t.Errorf("Expected 1.0 amplitude for 1.0 float, got %f", amp)
	}
}
