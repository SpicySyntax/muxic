# Usage Guide for Windows

This guide explains how to build and run the **muxic** multi-track audio recording CLI application on Windows.

## Prerequisites

### Install Zig

1. **Download Zig**: Visit the [official Zig downloads page](https://ziglang.org/download/) and download the latest version for Windows (e.g., `zig-x86_64-windows-0.16.0-dev.1484+d0ba6642b` or newer).

2. **Extract the Archive**: Extract the downloaded ZIP file to a location on your system, such as `C:\zig`.

3. **Add Zig to PATH**:
   - Open **System Properties** → **Environment Variables**
   - Under **System Variables**, find and select **Path**, then click **Edit**
   - Click **New** and add the path to your Zig installation (e.g., `C:\zig`)
   - Click **OK** to save

4. **Verify Installation**: Open a new PowerShell or Command Prompt window and run:
   ```powershell
   zig version
   ```
   You should see the Zig version number displayed.

## Building the Application

### Build for Development

To build the application in debug mode:

```powershell
zig build
```

This will compile the application and place the executable in `zig-out\bin\muxic.exe`.

### Build for Release

To build an optimized release version:

```powershell
zig build -Doptimize=ReleaseFast
```

Or for a smaller binary with debug symbols:

```powershell
zig build -Doptimize=ReleaseSafe
```

## Running the Application

### Quick Start

After building, you can run muxic using:

```powershell
zig build run -- <command> [arguments]
```

Or run the executable directly:

```powershell
.\zig-out\bin\muxic.exe <command> [arguments]
```

### Available Commands

#### Show Help

```powershell
zig build run -- help
```

#### Record a Track

Record a new audio track. You'll be prompted to press Enter to start recording, then Enter again to stop.

```powershell
zig build run -- record <track-name>
```

**Example:**
```powershell
zig build run -- record vocals
zig build run -- record guitar
zig build run -- record drums
```

#### List All Tracks

Display all recorded tracks with their file sizes:

```powershell
zig build run -- list
```

**Example output:**
```
📼 Recorded Tracks:
==================
  1. vocals (128 KB)
  2. guitar (256 KB)
  3. drums (192 KB)

Total: 3 track(s)
```

#### Play a Track

Play back a recorded track:

```powershell
zig build run -- play <track-name>
```

**Example:**
```powershell
zig build run -- play vocals
```

#### Mix Tracks

Mix all recorded tracks into a single output file:

```powershell
zig build run -- mix <output-name>
```

**Example:**
```powershell
zig build run -- mix final_mix
```

This will combine all tracks in the `tracks/` directory into a new file called `final_mix.wav`.

#### Export a Track

Export a track to a specific WAV file location:

```powershell
zig build run -- export <track-name> <output-file>
```

**Example:**
```powershell
zig build run -- export vocals my_vocals.wav
zig build run -- export guitar C:\Music\guitar_track.wav
```

## Workflow Example

Here's a typical workflow for creating a multi-track recording:

```powershell
# 1. Record your first track
zig build run -- record vocals

# 2. Record additional tracks
zig build run -- record guitar
zig build run -- record bass
zig build run -- record drums

# 3. List all tracks to verify
zig build run -- list

# 4. Play back individual tracks to check quality
zig build run -- play vocals
zig build run -- play guitar

# 5. Mix all tracks together
zig build run -- mix final_song

# 6. Export the final mix
zig build run -- export final_song my_song.wav
```

## Track Storage

All recorded tracks are stored in the `tracks/` directory as WAV files:
- **Format**: WAV (PCM)
- **Sample Rate**: 44100 Hz
- **Channels**: 2 (Stereo)
- **Bit Depth**: 16-bit

You can open these files in any audio software (Audacity, VLC, Windows Media Player, etc.).

## Troubleshooting

### "zig: command not found"

Make sure Zig is properly installed and added to your PATH. Close and reopen your terminal after adding Zig to PATH.

### Build Errors

If you encounter build errors, try:

1. Clean the build cache:
   ```powershell
   Remove-Item -Recurse -Force .zig-cache, zig-out
   zig build
   ```

2. Ensure you're using a compatible Zig version (0.15.2 or newer)

### Track Not Found

Make sure you're using the exact track name (without the `.wav` extension) when playing, mixing, or exporting tracks. Use `zig build run -- list` to see all available tracks.

## Development

### Run Tests

```powershell
zig build test
```

### Clean Build Artifacts

```powershell
Remove-Item -Recurse -Force .zig-cache, zig-out
```

## Next Steps

- Record multiple tracks sequentially
- Mix them together to create a complete song
- Export individual tracks or the final mix to share with others
- Use external audio software for advanced editing and effects
