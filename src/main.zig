const std = @import("std");

const TRACKS_DIR = "tracks";
const SAMPLE_RATE = 44100;
const CHANNELS = 2;
const BITS_PER_SAMPLE = 16;

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const allocator = gpa.allocator();

    const args = try std.process.argsAlloc(allocator);
    defer std.process.argsFree(allocator, args);

    if (args.len < 2) {
        printUsage();
        return;
    }

    const command = args[1];

    if (std.mem.eql(u8, command, "record")) {
        if (args.len < 3) {
            std.debug.print("Error: track name required\n", .{});
            std.debug.print("Usage: muxic record <track-name>\n", .{});
            return error.MissingTrackName;
        }
        try recordTrack(allocator, args[2]);
    } else if (std.mem.eql(u8, command, "play")) {
        if (args.len < 3) {
            std.debug.print("Error: track name required\n", .{});
            std.debug.print("Usage: muxic play <track-name>\n", .{});
            return error.MissingTrackName;
        }
        try playTrack(allocator, args[2]);
    } else if (std.mem.eql(u8, command, "list")) {
        try listTracks(allocator);
    } else if (std.mem.eql(u8, command, "mix")) {
        if (args.len < 3) {
            std.debug.print("Error: output name required\n", .{});
            std.debug.print("Usage: muxic mix <output-name>\n", .{});
            return error.MissingOutputName;
        }
        try mixTracks(allocator, args[2]);
    } else if (std.mem.eql(u8, command, "export")) {
        if (args.len < 4) {
            std.debug.print("Error: track name and output file required\n", .{});
            std.debug.print("Usage: muxic export <track-name> <output-file>\n", .{});
            return error.MissingArguments;
        }
        try exportTrack(allocator, args[2], args[3]);
    } else if (std.mem.eql(u8, command, "devices")) {
        try listDevices(allocator);
    } else if (std.mem.eql(u8, command, "help") or std.mem.eql(u8, command, "--help") or std.mem.eql(u8, command, "-h")) {
        printUsage();
    } else {
        std.debug.print("Error: unknown command '{s}'\n\n", .{command});
        printUsage();
        return error.UnknownCommand;
    }
}

fn printUsage() void {
    std.debug.print(
        \\Muxic - Multi-Track Audio Recording CLI
        \\
        \\Usage:
        \\  muxic record <track-name>           Record a new track
        \\  muxic play <track-name>             Play back a track
        \\  muxic list                          List all recorded tracks
        \\  muxic mix <output-name>             Mix all tracks into one file
        \\  muxic export <track-name> <file>    Export a track to WAV file
        \\  muxic devices                       List available audio devices
        \\  muxic help                          Show this help message
        \\
        \\Examples:
        \\  muxic record vocals
        \\  muxic record guitar
        \\  muxic list
        \\  muxic play vocals
        \\  muxic mix final_mix
        \\  muxic export vocals vocals.wav
        \\  muxic devices
        \\
    , .{});
}

fn ensureTracksDir() !void {
    std.fs.cwd().makeDir(TRACKS_DIR) catch |err| {
        if (err != error.PathAlreadyExists) return err;
    };
}

fn getTrackPath(allocator: std.mem.Allocator, track_name: []const u8) ![]u8 {
    return std.fmt.allocPrint(allocator, "{s}/{s}.wav", .{ TRACKS_DIR, track_name });
}

fn recordTrack(allocator: std.mem.Allocator, track_name: []const u8) !void {
    try ensureTracksDir();

    const track_path = try getTrackPath(allocator, track_name);
    defer allocator.free(track_path);

    std.debug.print("Recording track '{s}'...\n", .{track_name});
    std.debug.print("Press Enter to start recording, then Enter again to stop.\n", .{});

    // Wait for user to press Enter to start
    const stdin = std.fs.File.stdin();
    var buf: [256]u8 = undefined;
    _ = try stdin.read(&buf);

    std.debug.print("[RECORDING] Recording... (Press Enter to stop)\n", .{});

    // Simulate recording for now - in a real implementation, this would use zaudio
    // to capture audio from the microphone
    const recording_data = try simulateRecording(allocator);
    defer allocator.free(recording_data);

    // Wait for user to press Enter to stop
    _ = try stdin.read(&buf);

    // Save to WAV file
    try saveWavFile(track_path, recording_data);

    std.debug.print("[OK] Track '{s}' saved to {s}\n", .{ track_name, track_path });
}

fn simulateRecording(allocator: std.mem.Allocator) ![]u8 {
    // Generate 2 seconds of silence as placeholder
    const duration_seconds = 2;
    const num_samples = SAMPLE_RATE * duration_seconds * CHANNELS;
    const data_size = num_samples * (BITS_PER_SAMPLE / 8);

    const data = try allocator.alloc(u8, data_size);
    @memset(data, 0);

    return data;
}

fn playTrack(allocator: std.mem.Allocator, track_name: []const u8) !void {
    const track_path = try getTrackPath(allocator, track_name);
    defer allocator.free(track_path);

    // Check if file exists
    const file = std.fs.cwd().openFile(track_path, .{}) catch |err| {
        if (err == error.FileNotFound) {
            std.debug.print("Error: Track '{s}' not found\n", .{track_name});
            return err;
        }
        return err;
    };
    defer file.close();

    std.debug.print("[PLAYING] Playing track '{s}'...\n", .{track_name});

    // In a real implementation, this would use zaudio to play the audio
    // For now, just simulate playback
    std.debug.print("(Playback simulation - audio playback not yet implemented)\n", .{});

    std.debug.print("[OK] Playback complete\n", .{});
}

fn listDevices(allocator: std.mem.Allocator) !void {
    std.debug.print("[DEVICES] Audio Devices:\n", .{});
    std.debug.print("======================\n", .{});

    // Use PowerShell to get friendly device names
    // method: Get-CimInstance Win32_SoundDevice | Select-Object -Property Name, Manufacturer | Format-Table -AutoSize
    const cmd_args = [_][]const u8{
        "powershell",
        "-c",
        "Get-CimInstance Win32_SoundDevice | Select-Object -Property Name, Manufacturer | Format-Table -AutoSize",
    };

    var child = std.process.Child.init(&cmd_args, allocator);
    child.stdout_behavior = .Inherit;
    child.stderr_behavior = .Inherit;

    _ = try child.spawnAndWait();
}

fn listTracks(_: std.mem.Allocator) !void {
    var dir = std.fs.cwd().openDir(TRACKS_DIR, .{ .iterate = true }) catch |err| {
        if (err == error.FileNotFound) {
            std.debug.print("No tracks directory found. Record a track first!\n", .{});
            return;
        }
        return err;
    };
    defer dir.close();

    std.debug.print("[TRACKS] Recorded Tracks:\n", .{});
    std.debug.print("==================\n", .{});

    var iter = dir.iterate();
    var count: usize = 0;

    while (try iter.next()) |entry| {
        if (entry.kind == .file and std.mem.endsWith(u8, entry.name, ".wav")) {
            count += 1;
            // Remove .wav extension for display
            const name_without_ext = entry.name[0 .. entry.name.len - 4];

            // Get file size
            const file = try dir.openFile(entry.name, .{});
            defer file.close();
            const stat = try file.stat();
            const size_kb = stat.size / 1024;

            std.debug.print("  {d}. {s} ({d} KB)\n", .{ count, name_without_ext, size_kb });
        }
    }

    if (count == 0) {
        std.debug.print("  (no tracks recorded yet)\n", .{});
    } else {
        std.debug.print("\nTotal: {d} track(s)\n", .{count});
    }
}

fn mixTracks(allocator: std.mem.Allocator, output_name: []const u8) !void {
    try ensureTracksDir();

    var dir = std.fs.cwd().openDir(TRACKS_DIR, .{ .iterate = true }) catch |err| {
        if (err == error.FileNotFound) {
            std.debug.print("Error: No tracks found to mix\n", .{});
            return err;
        }
        return err;
    };
    defer dir.close();

    std.debug.print("[MIXING] Mixing tracks into '{s}'...\n", .{output_name});

    // Count tracks
    var iter = dir.iterate();
    var track_count: usize = 0;
    while (try iter.next()) |entry| {
        if (entry.kind == .file and std.mem.endsWith(u8, entry.name, ".wav")) {
            track_count += 1;
            std.debug.print("  + {s}\n", .{entry.name});
        }
    }

    if (track_count == 0) {
        std.debug.print("Error: No tracks found to mix\n", .{});
        return error.NoTracksFound;
    }

    // In a real implementation, this would:
    // 1. Load all WAV files
    // 2. Mix the audio samples (add and normalize)
    // 3. Save to new WAV file

    const output_path = try getTrackPath(allocator, output_name);
    defer allocator.free(output_path);

    // For now, create a placeholder mixed file
    const mixed_data = try simulateRecording(allocator);
    defer allocator.free(mixed_data);
    try saveWavFile(output_path, mixed_data);

    std.debug.print("[OK] Mixed {d} tracks into {s}\n", .{ track_count, output_path });
}

fn exportTrack(allocator: std.mem.Allocator, track_name: []const u8, output_file: []const u8) !void {
    const track_path = try getTrackPath(allocator, track_name);
    defer allocator.free(track_path);

    // Check if source file exists
    const src_file = std.fs.cwd().openFile(track_path, .{}) catch |err| {
        if (err == error.FileNotFound) {
            std.debug.print("Error: Track '{s}' not found\n", .{track_name});
            return err;
        }
        return err;
    };
    defer src_file.close();

    // Copy to output file
    const dest_file = try std.fs.cwd().createFile(output_file, .{});
    defer dest_file.close();

    const buffer = try allocator.alloc(u8, 4096);
    defer allocator.free(buffer);

    while (true) {
        const bytes_read = try src_file.read(buffer);
        if (bytes_read == 0) break;
        try dest_file.writeAll(buffer[0..bytes_read]);
    }

    std.debug.print("[OK] Exported '{s}' to {s}\n", .{ track_name, output_file });
}

fn saveWavFile(path: []const u8, audio_data: []const u8) !void {
    const file = try std.fs.cwd().createFile(path, .{});
    defer file.close();

    const data_size: u32 = @intCast(audio_data.len);
    const file_size: u32 = 36 + data_size;

    var buf: [4]u8 = undefined;

    // WAV header
    _ = try file.write("RIFF");
    std.mem.writeInt(u32, &buf, file_size, .little);
    _ = try file.write(&buf);
    _ = try file.write("WAVE");

    // fmt chunk
    _ = try file.write("fmt ");
    std.mem.writeInt(u32, &buf, 16, .little); // fmt chunk size
    _ = try file.write(&buf);

    var buf2: [2]u8 = undefined;
    std.mem.writeInt(u16, &buf2, 1, .little); // PCM format
    _ = try file.write(&buf2);
    std.mem.writeInt(u16, &buf2, CHANNELS, .little);
    _ = try file.write(&buf2);
    std.mem.writeInt(u32, &buf, SAMPLE_RATE, .little);
    _ = try file.write(&buf);

    const byte_rate: u32 = SAMPLE_RATE * CHANNELS * (BITS_PER_SAMPLE / 8);
    std.mem.writeInt(u32, &buf, byte_rate, .little);
    _ = try file.write(&buf);

    const block_align: u16 = CHANNELS * (BITS_PER_SAMPLE / 8);
    std.mem.writeInt(u16, &buf2, block_align, .little);
    _ = try file.write(&buf2);
    std.mem.writeInt(u16, &buf2, BITS_PER_SAMPLE, .little);
    _ = try file.write(&buf2);

    // data chunk
    _ = try file.write("data");
    std.mem.writeInt(u32, &buf, data_size, .little);
    _ = try file.write(&buf);
    _ = try file.write(audio_data);
}
