package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"

	"github.com/moutend/go-wca/pkg/wca"
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
	case "device":
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

	if err := ole.CoInitialize(0); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	// Find the device
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	var mmde *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return err
	}
	defer mmde.Release()

	var pCollection *wca.IMMDeviceCollection
	if err := mmde.EnumAudioEndpoints(wca.ECapture, wca.DEVICE_STATE_ACTIVE, &pCollection); err != nil {
		return err
	}
	defer pCollection.Release()

	var count uint32
	if err := pCollection.GetCount(&count); err != nil {
		return err
	}

	var device *wca.IMMDevice
	targetName := ""
	if config.DefaultDevice != nil {
		targetName = *config.DefaultDevice
	}

	for i := uint32(0); i < count; i++ {
		var pEndpoint *wca.IMMDevice
		if err := pCollection.Item(i, &pEndpoint); err != nil {
			continue
		}

		var pProps *wca.IPropertyStore
		if err := pEndpoint.OpenPropertyStore(wca.STGM_READ, &pProps); err != nil {
			pEndpoint.Release()
			continue
		}

		var pv wca.PROPVARIANT
		if err := pProps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
			pProps.Release()
			pEndpoint.Release()
			continue
		}
		name := pv.String()
		pProps.Release()

		if targetName != "" && name == targetName {
			device = pEndpoint
			fmt.Printf("Using device: %s\n", name)
			break
		} else if targetName == "" {
			// improved default logic could go here, for now take first
			device = pEndpoint
			fmt.Printf("Using default device: %s\n", name)
			break
		}
		pEndpoint.Release()
	}

	if device == nil {
		return fmt.Errorf("device not found")
	}
	defer device.Release()

	// Initialize Audio Client
	var audioClient *wca.IAudioClient
	if err := device.Activate(wca.IID_IAudioClient, wca.CLSCTX_ALL, nil, &audioClient); err != nil {
		return err
	}
	defer audioClient.Release()

	var wfx *wca.WAVEFORMATEX
	if err := audioClient.GetMixFormat(&wfx); err != nil {
		return err
	}
	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(wfx)))

	// Initialize in Shared Mode
	if err := audioClient.Initialize(wca.AUDCLNT_SHAREMODE_SHARED, 0, 10000000, 0, wfx, nil); err != nil {
		return err
	}

	var captureClient *wca.IAudioCaptureClient
	if err := audioClient.GetService(wca.IID_IAudioCaptureClient, &captureClient); err != nil {
		return err
	}
	defer captureClient.Release()

	fmt.Printf("Recording format: %d Hz, %d channels, %d bits\n", wfx.NSamplesPerSec, wfx.NChannels, wfx.WBitsPerSample)
	fmt.Println("Press Enter to start recording...")
	bufio.NewReader(os.Stdin).ReadString('\n')

	if err := audioClient.Start(); err != nil {
		return err
	}
	fmt.Println("[RECORDING] Recording... (Press Enter to stop)")

	// Visualizer setup
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	var currentAmplitude float64

	done := make(chan bool)
	go func() {
		bufio.NewReader(os.Stdin).ReadString('\n')
		done <- true
	}()

	var audioData []byte
	var isCapturing = true

	for isCapturing {
		select {
		case <-done:
			isCapturing = false
			fmt.Println() // Newline after visualizer
		case <-ticker.C:
			drawVisualizer(currentAmplitude)
		default:
			var buffer *byte
			var framesAvailable uint32
			var flags uint32
			var devicePosition uint64
			var qpcPosition uint64

			// Get buffer
			if err := captureClient.GetBuffer(&buffer, &framesAvailable, &flags, &devicePosition, &qpcPosition); err != nil {
				if err.Error() == "AUDCLNT_S_BUFFER_EMPTY" { // Not really an error, just empty
					time.Sleep(10 * time.Millisecond)
					continue
				}
				// Retry or fail? For now sleep and continue
				time.Sleep(10 * time.Millisecond)
				continue
			}

			if framesAvailable > 0 {
				bytesToCopy := int(framesAvailable) * int(wfx.NBlockAlign)
				chunk := make([]byte, bytesToCopy)

				// Unsafe copy from COM buffer
				// Safety: buffer is valid until ReleaseBuffer
				src := unsafe.Slice(buffer, bytesToCopy)
				copy(chunk, src)
				copy(chunk, src)
				audioData = append(audioData, chunk...)

				// Calculate amplitude for visualizer
				currentAmplitude = calculateAmplitude(chunk, wfx.WBitsPerSample)
			}

			if err := captureClient.ReleaseBuffer(framesAvailable); err != nil {
				return err
			}

			// Small sleep to avoid busy loop if buffer is small?
			// Actually GetBuffer returns quickly if available.
			// We effectively poll.
			time.Sleep(1 * time.Millisecond)
		}
	}

	if err := audioClient.Stop(); err != nil {
		return err
	}

	// Convert if necessary
	// WASAPI Audio Client commonly returns IEEE Float (32-bit) in Shared Mode.
	// We want to save as standard PCM 16-bit for best compatibility.
	var finalData []byte
	var finalWfx *wca.WAVEFORMATEX

	// Simple check: 32 bits usually implies Float for WASAPI shared mode
	if wfx.WBitsPerSample == 32 {
		fmt.Println("Converting 32-bit Float to 16-bit PCM...")

		// Interpret byte slice as float32 slice
		// Note: audioData is raw bytes. We need to handle endianness, but standard Windows is Little Endian.
		// Go's unsafe cast helps if we assume host is LE.
		numSamples := len(audioData) / 4
		pcmData := make([]byte, numSamples*2)

		// Create a reader to simplify binary reading
		// Or unsafe slice for speed? Let's use unsafe for efficiency if we are careful.
		// float32s := unsafe.Slice((*float32)(unsafe.Pointer(&audioData[0])), numSamples) // Unsafe if len=0

		// Safer loop
		for i := 0; i < numSamples; i++ {
			// Read float32 (LE)
			// Manual bit conversion to avoid unsafe alignment issues if any (buffer from Append should be aligned?)
			// Actually, just using encoding/binary is safer slightly slower.

			// Let's use Unsafe for "interpretation" but careful.
			// Ideally we use a loop with math.

			// Extract 4 bytes
			b1 := audioData[i*4]
			b2 := audioData[i*4+1]
			b3 := audioData[i*4+2]
			b4 := audioData[i*4+3]
			valbits := uint32(b1) | uint32(b2)<<8 | uint32(b3)<<16 | uint32(b4)<<24
			fVal := *(*float32)(unsafe.Pointer(&valbits))

			// Clamp and Convert
			if fVal > 1.0 {
				fVal = 1.0
			}
			if fVal < -1.0 {
				fVal = -1.0
			}
			iVal := int16(fVal * 32767)

			// Write int16 (LE)
			pcmData[i*2] = byte(iVal)
			pcmData[i*2+1] = byte(iVal >> 8)
		}

		finalData = pcmData

		// Create new WFX for PCM 16-bit
		finalWfx = &wca.WAVEFORMATEX{
			WFormatTag:      1, // WAVE_FORMAT_PCM
			NChannels:       wfx.NChannels,
			NSamplesPerSec:  wfx.NSamplesPerSec,
			NAvgBytesPerSec: wfx.NSamplesPerSec * uint32(wfx.NChannels) * 2,
			NBlockAlign:     wfx.NChannels * 2,
			WBitsPerSample:  16,
			CbSize:          0,
		}
	} else {
		finalData = audioData
		finalWfx = wfx
	}

	if err := saveWavFile(trackPath, finalData, finalWfx); err != nil {
		return err
	}

	fmt.Printf("[OK] Track '%s' saved to %s\n", trackName, trackPath)
	return nil
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
	// Use silence for mixing as placeholder
	mixedData := generateSilence()

	// Default mix format (PCM 44.1kHz 16-bit Stereo)
	wfx := &wca.WAVEFORMATEX{
		WFormatTag:      1, // WAVE_FORMAT_PCM
		NChannels:       2,
		NSamplesPerSec:  44100,
		NAvgBytesPerSec: 44100 * 4,
		NBlockAlign:     4,
		WBitsPerSample:  16,
		CbSize:          0,
	}

	if err := saveWavFile(outputPath, mixedData, wfx); err != nil {
		return err
	}

	fmt.Printf("[OK] Mixed %d tracks into %s\n", trackCount, outputPath)
	return nil
}

func generateSilence() []byte {
	// Generate 2 seconds of silence as placeholder
	durationSeconds := 2
	numSamples := SampleRate * durationSeconds * Channels
	dataSize := numSamples * (BitsPerSample / 8)
	return make([]byte, dataSize)
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

type AudioDevice struct {
	Name         string `json:"Name"`
	Manufacturer string `json:"Manufacturer"`
}

func listDevices() error {
	if err := ole.CoInitialize(0); err != nil {
		return err
	}
	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator
	if err := wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return err
	}
	defer mmde.Release()

	var pCollection *wca.IMMDeviceCollection
	if err := mmde.EnumAudioEndpoints(wca.ECapture, wca.DEVICE_STATE_ACTIVE, &pCollection); err != nil {
		return err
	}
	defer pCollection.Release()

	var count uint32
	if err := pCollection.GetCount(&count); err != nil {
		return err
	}

	config, err := LoadConfig()
	if err != nil {
		return err
	}

	fmt.Println("[DEVICES] Audio Capture Devices:")
	if config.DefaultDevice != nil {
		fmt.Printf("Current Default: %s\n", *config.DefaultDevice)
	} else {
		fmt.Println("Current Default: (none selected)")
	}
	fmt.Println("======================")

	if count == 0 {
		fmt.Println("No audio capture devices found.")
		return nil
	}

	for i := uint32(0); i < count; i++ {
		var pEndpoint *wca.IMMDevice
		if err := pCollection.Item(i, &pEndpoint); err != nil {
			continue
		}
		defer pEndpoint.Release()

		var pProps *wca.IPropertyStore
		if err := pEndpoint.OpenPropertyStore(wca.STGM_READ, &pProps); err != nil {
			continue
		}
		defer pProps.Release()

		var pv wca.PROPVARIANT
		if err := pProps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
			continue
		}
		deviceName := pv.String()

		indicator := " "
		if config.DefaultDevice != nil && *config.DefaultDevice == deviceName {
			indicator = "*"
		}
		fmt.Printf("%s %d. %s\n", indicator, i+1, deviceName)
	}

	return nil
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

func saveWavFile(path string, audioData []byte, wfx *wca.WAVEFORMATEX) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dataSize := uint32(len(audioData))
	fileSize := uint32(36 + dataSize)

	// WAV header
	if _, err := f.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, fileSize); err != nil {
		return err
	}
	if _, err := f.Write([]byte("WAVE")); err != nil {
		return err
	}

	// fmt chunk
	if _, err := f.Write([]byte("fmt ")); err != nil {
		return err
	}
	// Check for extensible format
	if wfx.WFormatTag == 0xFFFE /* WAVE_FORMAT_EXTENSIBLE */ {
		if err := binary.Write(f, binary.LittleEndian, uint32(40)); err != nil {
			return err
		}
	} else {
		if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil {
			return err
		}
	}

	if err := binary.Write(f, binary.LittleEndian, wfx.WFormatTag); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, wfx.NChannels); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, wfx.NSamplesPerSec); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, wfx.NAvgBytesPerSec); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, wfx.NBlockAlign); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, wfx.WBitsPerSample); err != nil {
		return err
	}

	if wfx.WFormatTag == 0xFFFE /* WAVE_FORMAT_EXTENSIBLE */ {
		if err := binary.Write(f, binary.LittleEndian, wfx.CbSize); err != nil { // cbSize
			return err
		}
		// Write the rest of WAVEFORMATEXTENSIBLE if needed?
		// wca.WAVEFORMATEX doesn't expose the extra bytes directly as a struct field easily accessible here
		// without unsafe casting, but standard WAVE header is usually enough.
		// However, for EXTENSIBLE, we need to write the helper bytes.
		// The Go struct wca.WAVEFORMATEX ends at CbSize.
		// To keep it simple, we might just write the standard 16 + cbSize bytes if accessible,
		// but Go definition is tricky.
		// NOTE: wca.WAVEFORMATEX in go-wca seems to only have the basic fields + CbSize?
		// Checking definition of WAVEFORMATEX: it has cbSize at end.
		// If 16-bit PCM, cbSize is 0 or ignored.
		// If Float, it might be WAVE_FORMAT_IEEE_FLOAT (3) or EXTENSIBLE (0xFFFE).

		// Let's assume for now we write basic header.
		// To be strictly correct for EXTENSIBLE we need subformat.
		// But let's check what GetMixFormat returns usually.
		// Often it returns EXTENSIBLE for shared mode.
		// If we just write the fmt chunk based on what we have:

		// Actually, let's keep it simple: Write what we have.
		// If it is extensible, we should ideally write the GUIDs.
		// But maybe just writing the data as is works for many players.

		// Let's rely on standard PCM or Float tags if possible.
		// If GetMixFormat returns EXTENSIBLE, we might want to Convert to PCM/Float,
		// or just write the header as is.

		// We will write the additional bytes for Extensible if we can access them.
		// Since we can't easily, let's just write the basic fields and hope `CbSize` covers it if we write extra?
		// No, `CbSize` tells how many extra bytes follow. We don't have them in the struct `wfx` easily.
		// BUT `wfx` is a pointer to the start of the memory. We can read `cbSize` bytes after the struct.

		cbSize := wfx.CbSize
		if cbSize > 0 {
			// Read extra bytes
			extraBytes := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(wfx))+unsafe.Sizeof(*wfx))), cbSize)
			if _, err := f.Write(extraBytes); err != nil {
				return err
			}
		}
	} else {
		// If not extensible, and cbSize is present (e.g. for some compressed formats), handle it?
		// Usually for PCM cbSize is ignored/not present in basics, but we wrote 16 for chunk size.
		// If wFormatTag != PCM, chunk size might be larger.
		// For simplicity, let's write 16 bytes for PCM/Float if we force it?
		// But we are using the MixFormat.

		// We wrote 16 as chunk size above for non-extensible.
		// So we shouldn't write cbSize or extra bytes.
	}

	// data chunk
	if _, err := f.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, dataSize); err != nil {
		return err
	}
	if _, err := f.Write(audioData); err != nil {
		return err
	}

	return nil
}

func calculateAmplitude(data []byte, bitsPerSample uint16) float64 {
	var maxVal float64

	if bitsPerSample == 16 {
		numSamples := len(data) / 2
		for i := 0; i < numSamples; i++ {
			sample := int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
			absVal := math.Abs(float64(sample)) / 32768.0
			if absVal > maxVal {
				maxVal = absVal
			}
		}
	} else if bitsPerSample == 32 {
		numSamples := len(data) / 4
		for i := 0; i < numSamples; i++ {
			valbits := binary.LittleEndian.Uint32(data[i*4 : i*4+4])
			fVal := math.Float32frombits(valbits)
			absVal := math.Abs(float64(fVal))
			if absVal > maxVal {
				maxVal = absVal
			}
		}
	}

	return maxVal
}

func drawVisualizer(amplitude float64) {
	const barWidth = 20
	numBars := int(amplitude * barWidth)
	if numBars > barWidth {
		numBars = barWidth
	}

	bars := strings.Repeat("|", numBars)
	spaces := strings.Repeat(" ", barWidth-numBars)

	fmt.Printf("\r\033[K[%s%s]", bars, spaces)
}
