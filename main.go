package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"bufio"
)

const (
	TracksDir     = "tracks"
	SampleRate    = 44100
	Channels      = 2
	BitsPerSample = 16
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	args := os.Args[1:] // args excluding program name

	var err error
	switch command {
	case "record":
		if len(args) < 2 {
			fmt.Println("Error: track name required")
			fmt.Println("Usage: muxic record <track-name>")
			os.Exit(1)
		}
		err = recordTrack(args[1])
	case "play":
		if len(args) < 2 {
			fmt.Println("Error: track name required")
			fmt.Println("Usage: muxic play <track-name>")
			os.Exit(1)
		}
		err = playTrack(args[1])
	case "list":
		err = listTracks()
	case "mix":
		if len(args) < 2 {
			fmt.Println("Error: output name required")
			fmt.Println("Usage: muxic mix <output-name>")
			os.Exit(1)
		}
		err = mixTracks(args[1])
	case "export":
		if len(args) < 3 {
			fmt.Println("Error: track name and output file required")
			fmt.Println("Usage: muxic export <track-name> <output-file>")
			os.Exit(1)
		}
		err = exportTrack(args[1], args[2])
	case "devices", "device":
		if len(args) >= 2 && args[1] == "select" {
			if len(args) < 3 {
				fmt.Println("Error: device name required")
				fmt.Println("Usage: muxic device select <device-name>")
				os.Exit(1)
			}
			deviceName := strings.Join(args[2:], " ")
			err = selectDevice(deviceName)
		} else {
			if len(args) >= 2 && args[1] == "list" {
				err = listDevices()
			} else {
				// Default to list
				err = listDevices()
			}
		}
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Error: unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`Muxic - Multi-Track Audio Recording CLI

Usage:
  muxic record <track-name>           Record a new track
  muxic play <track-name>             Play back a track
  muxic list                          List all recorded tracks
  muxic mix <output-name>             Mix all tracks into one file
  muxic export <track-name> <file>    Export a track to WAV file
  muxic device list                   List available audio devices
  muxic device select <name>          Select default recording device
  muxic help                          Show this help message

Examples:
  muxic record vocals
  muxic record guitar
  muxic list
  muxic play vocals
  muxic mix final_mix
  muxic export vocals vocals.wav
  muxic devices
`)
}

func ensureTracksDir() error {
	return os.MkdirAll(TracksDir, 0755)
}

func getTrackPath(trackName string) string {
	return filepath.Join(TracksDir, trackName+".wav")
}

func recordTrack(trackName string) error {
	if err := ensureTracksDir(); err != nil {
		return err
	}

	trackPath := getTrackPath(trackName)

	config, err := LoadConfig()
	if err != nil {
		return err
	}

	deviceName := "Default Input"
	if config.DefaultDevice != nil {
		deviceName = *config.DefaultDevice
	}

	fmt.Printf("Recording track '%s' using device: '%s'...\n", trackName, deviceName)
	fmt.Println("Press Enter to start recording, then Enter again to stop.")

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	fmt.Println("[RECORDING] Recording... (Press Enter to stop)")
	
	recordingData := simulateRecording()
	
	reader.ReadString('\n')

	if err := saveWavFile(trackPath, recordingData); err != nil {
		return err
	}

	fmt.Printf("[OK] Track '%s' saved to %s\n", trackName, trackPath)
	return nil
}

func simulateRecording() []byte {
	// Generate 2 seconds of silence as placeholder
	durationSeconds := 2
	numSamples := SampleRate * durationSeconds * Channels
	dataSize := numSamples * (BitsPerSample / 8)
	return make([]byte, dataSize)
}

func playTrack(trackName string) error {
	trackPath := getTrackPath(trackName)

	if _, err := os.Stat(trackPath); os.IsNotExist(err) {
		return fmt.Errorf("Track '%s' not found", trackName)
	}

	fmt.Printf("[PLAYING] Playing track '%s'...\n", trackName)
	fmt.Println("(Playback simulation - audio playback not yet implemented)")
	fmt.Println("[OK] Playback complete")
	return nil
}

func listTracks() error {
	entries, err := os.ReadDir(TracksDir)
	if os.IsNotExist(err) {
		fmt.Println("No tracks directory found. Record a track first!")
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Println("[TRACKS] Recorded Tracks:")
	fmt.Println("==================")

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".wav") {
			count++
			nameWithoutExt := strings.TrimSuffix(entry.Name(), ".wav")
			
			info, err := entry.Info()
			if err != nil {
				continue
			}
			sizeKb := info.Size() / 1024
			
			fmt.Printf("  %d. %s (%d KB)\n", count, nameWithoutExt, sizeKb)
		}
	}

	if count == 0 {
		fmt.Println("  (no tracks recorded yet)")
	} else {
		fmt.Printf("\nTotal: %d track(s)\n", count)
	}
	return nil
}

func mixTracks(outputName string) error {
	if err := ensureTracksDir(); err != nil {
		return err
	}

	entries, err := os.ReadDir(TracksDir)
	if os.IsNotExist(err) {
		fmt.Println("Error: No tracks found to mix")
		return err
	}
	if err != nil {
		return err
	}

	fmt.Printf("[MIXING] Mixing tracks into '%s'...\n", outputName)

	trackCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".wav") {
			trackCount++
			fmt.Printf("  + %s\n", entry.Name())
		}
	}

	if trackCount == 0 {
		return fmt.Errorf("No tracks found to mix")
	}

	outputPath := getTrackPath(outputName)
	mixedData := simulateRecording()
	
	if err := saveWavFile(outputPath, mixedData); err != nil {
		return err
	}

	fmt.Printf("[OK] Mixed %d tracks into %s\n", trackCount, outputPath)
	return nil
}

func exportTrack(trackName, outputFile string) error {
	trackPath := getTrackPath(trackName)

	srcFile, err := os.Open(trackPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("Track '%s' not found", trackName)
	}
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return err
	}

	fmt.Printf("[OK] Exported '%s' to %s\n", trackName, outputFile)
	return nil
}

func listDevices() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	fmt.Println("[DEVICES] Audio Devices:")
	if config.DefaultDevice != nil {
		fmt.Printf("Current Default: %s\n", *config.DefaultDevice)
	} else {
		fmt.Println("Current Default: (none selected)")
	}
	fmt.Println("======================")
	fmt.Println("======================")

	cmd := exec.Command("powershell", "-c", "Get-CimInstance Win32_SoundDevice | Select-Object -Property Name, Manufacturer | Format-Table -AutoSize")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func selectDevice(deviceName string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.DefaultDevice = &deviceName
	if err := config.Save(); err != nil {
		return err
	}

	fmt.Printf("Selected default recording device: '%s'\n", deviceName)
	return nil
}

func saveWavFile(path string, audioData []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dataSize := uint32(len(audioData))
	fileSize := uint32(36 + dataSize)

	// WAV header
	if _, err := f.Write([]byte("RIFF")); err != nil { return err }
	if err := binary.Write(f, binary.LittleEndian, fileSize); err != nil { return err }
	if _, err := f.Write([]byte("WAVE")); err != nil { return err }

	// fmt chunk
	if _, err := f.Write([]byte("fmt ")); err != nil { return err }
	if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil { return err } // Chunk size
	if err := binary.Write(f, binary.LittleEndian, uint16(1)); err != nil { return err } // PCM
	if err := binary.Write(f, binary.LittleEndian, uint16(Channels)); err != nil { return err }
	if err := binary.Write(f, binary.LittleEndian, uint32(SampleRate)); err != nil { return err }

	byteRate := uint32(SampleRate * Channels * (BitsPerSample / 8))
	if err := binary.Write(f, binary.LittleEndian, byteRate); err != nil { return err }

	blockAlign := uint16(Channels * (BitsPerSample / 8))
	if err := binary.Write(f, binary.LittleEndian, blockAlign); err != nil { return err }
	if err := binary.Write(f, binary.LittleEndian, uint16(BitsPerSample)); err != nil { return err }

	// data chunk
	if _, err := f.Write([]byte("data")); err != nil { return err }
	if err := binary.Write(f, binary.LittleEndian, dataSize); err != nil { return err }
	if _, err := f.Write(audioData); err != nil { return err }

	return nil
}
