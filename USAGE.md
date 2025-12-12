# Usage Guide for Windows

This guide explains how to build and run the **muxic** multi-track audio recording CLI application on Windows.

## Prerequisites

### Install Go

1. **Download Go**: Visit the [official Go downloads page](https://go.dev/dl/) and download the latest version for Windows.

2. **Install**: Run the installer and follow the prompts.

3. **Verify Installation**: Open a new PowerShell or Command Prompt window and run:
   ```powershell
   go version
   ```
   You should see the Go version number displayed.

## Building the Application

### Build Executable

To build the application:

```powershell
go build -o muxic.exe
```

This will create `muxic.exe` in the current directory.

## Running the Application

### Development Mode

You can run the application directly without building the binary first:

```powershell
go run . <command> [arguments]
```

### Production Mode

After building, run the executable:

```powershell
.\muxic.exe <command> [arguments]
```

### Available Commands

#### Show Help

```powershell
.\muxic.exe help
```

#### Record a Track

Record a new audio track. You'll be prompted to press Enter to start recording, then Enter again to stop.

```powershell
.\muxic.exe record <track-name>
```

**Example:**
```powershell
.\muxic.exe record vocals
.\muxic.exe record guitar
.\muxic.exe record drums
```

#### List All Tracks

Display all recorded tracks with their file sizes:

```powershell
.\muxic.exe list
```

**Example output:**
```
[TRACKS] Recorded Tracks:
==================
  1. vocals (128 KB)
  2. guitar (256 KB)
  3. drums (192 KB)

Total: 3 track(s)
```

#### Play a Track

Play back a recorded track:

```powershell
.\muxic.exe play <track-name>
```

**Example:**
```powershell
.\muxic.exe play vocals
```

#### Mix Tracks

Mix all recorded tracks into a single output file:

```powershell
.\muxic.exe mix <output-name>
```

**Example:**
```powershell
.\muxic.exe mix final_mix
```

This will combine all tracks in the `tracks/` directory into a new file called `final_mix.wav`.

#### Export a Track

Export a track to a specific WAV file location:

```powershell
.\muxic.exe export <track-name> <output-file>
```

**Example:**
```powershell
.\muxic.exe export vocals my_vocals.wav
.\muxic.exe export guitar C:\Music\guitar_track.wav
```

## Workflow Example

Here's a typical workflow for creating a multi-track recording:

```powershell
# 1. Record your first track
.\muxic.exe record vocals

# 2. Record additional tracks
.\muxic.exe record guitar
.\muxic.exe record bass
.\muxic.exe record drums

# 3. List all tracks to verify
.\muxic.exe list

# 4. Play back individual tracks to check quality
.\muxic.exe play vocals
.\muxic.exe play guitar

# 5. Mix all tracks together
.\muxic.exe mix final_song

# 6. Export the final mix
.\muxic.exe export final_song my_song.wav
```

## Track Storage

All recorded tracks are stored in the `tracks/` directory as WAV files:
- **Format**: WAV (PCM)
- **Sample Rate**: 44100 Hz
- **Channels**: 2 (Stereo)
- **Bit Depth**: 16-bit

You can open these files in any audio software (Audacity, VLC, Windows Media Player, etc.).

## Troubleshooting

### "go: command not found"

Make sure Go is properly installed and added to your PATH. Close and reopen your terminal after installation.

### Track Not Found

Make sure you're using the exact track name (without the `.wav` extension) when playing, mixing, or exporting tracks. Use `.\muxic.exe list` to see all available tracks.

## Development

### Run Tests

```powershell
go test ./...
```

### Clean Build Artifacts

```powershell
Remove-Item muxic.exe
```

## Next Steps

- Record multiple tracks sequentially
- Mix them together to create a complete song
- Export individual tracks or the final mix to share with others
- Use external audio software for advanced editing and effects
