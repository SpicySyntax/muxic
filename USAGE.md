# Usage Guide for Windows

This guide explains how to build and run the **muxic** application on Windows.

## Prerequisites

### Install Zig

1. **Download Zig**: Visit the [official Zig downloads page](https://ziglang.org/download/) and download the latest version for Windows (e.g., `zig-windows-x86_64-0.13.0.zip` or newer).

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

Or for the smallest binary size:

```powershell
zig build -Doptimize=ReleaseSmall
```

## Running the Application

### Run Directly

You can build and run the application in one command:

```powershell
zig build run
```

### Run with Arguments

To pass arguments to the application:

```powershell
zig build run -- [your arguments here]
```

### Run the Compiled Executable

After building, you can run the executable directly:

```powershell
.\zig-out\bin\muxic.exe
```

## Testing

To run the unit tests:

```powershell
zig build test
```

## Cleaning Build Artifacts

To clean the build cache and output:

```powershell
Remove-Item -Recurse -Force .zig-cache, zig-out
```

## Troubleshooting

### "zig is not recognized"

If you get this error, Zig is not in your PATH. Make sure you:
1. Added Zig to your PATH environment variable
2. Opened a **new** terminal window after modifying PATH
3. Verified the path is correct

### Build Errors

If you encounter build errors:
1. Ensure you're using a compatible Zig version (check `build.zig.zon` for requirements)
2. Try cleaning the build cache: `Remove-Item -Recurse -Force .zig-cache`
3. Check that all dependencies are properly configured in `build.zig.zon`

### Permission Issues

If you encounter permission errors when running the executable:
1. Make sure you have write permissions in the project directory
2. Try running PowerShell as Administrator
3. Check your antivirus software isn't blocking the executable

## Additional Build Options

### View All Build Steps

To see all available build steps:

```powershell
zig build --help
```

### Specify Target Platform

To cross-compile for a different platform:

```powershell
zig build -Dtarget=x86_64-windows
```

## Development Workflow

A typical development workflow on Windows:

1. Make changes to source files in `src/`
2. Build and run: `zig build run`
3. Run tests: `zig build test`
4. Build release version when ready: `zig build -Doptimize=ReleaseFast`

## Project Structure

```
muxic/
├── build.zig          # Build configuration
├── build.zig.zon      # Dependencies and package info
├── src/               # Source files
│   └── main.zig       # Main entry point
├── zig-out/           # Build output (generated)
│   └── bin/
│       └── muxic.exe  # Compiled executable
└── .zig-cache/        # Build cache (generated)
```
