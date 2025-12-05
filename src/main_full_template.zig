// Full implementation template for Muxic Audio Recording App
// This file contains the complete implementation that requires external dependencies
// Once you add Capy UI and zaudio dependencies, you can use this code in main.zig

const std = @import("std");
const capy = @import("capy");
const zaudio = @import("zaudio");

var is_recording = false;
var audio_engine: ?*zaudio.Engine = null;
var audio_device: ?*zaudio.Device = null;
var recording_buffer: std.ArrayList(u8) = undefined;
var allocator: std.mem.Allocator = undefined;

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    allocator = gpa.allocator();

    recording_buffer = std.ArrayList(u8).init(allocator);
    defer recording_buffer.deinit();

    try capy.backend.init();
    defer capy.backend.deinit();

    var window = try capy.Window.init();
    defer window.deinit();

    try window.set(
        capy.column(.{ .spacing = 10 }, .{
            capy.row(.{ .spacing = 5 }, .{
                capy.button(.{
                    .label = "Start Recording",
                    .onclick = onRecordButtonClick,
                }),
            }),
            capy.label(.{
                .text = "Status: Idle",
            }),
        }),
    );

    window.setTitle("Muxic - Audio Recorder");
    window.setPreferredSize(400, 300);
    window.show();

    capy.runEventLoop();
}

fn onRecordButtonClick(button: *capy.Button) !void {
    if (is_recording) {
        try stopRecording();
        button.setLabel("Start Recording");
    } else {
        try startRecording();
        button.setLabel("Stop Recording");
    }
}

fn startRecording() !void {
    std.debug.print("Starting recording...\n", .{});
    
    if (audio_engine == null) {
        audio_engine = try zaudio.Engine.create(allocator);
    }

    var device_config = zaudio.DeviceConfig.init(.capture);
    device_config.capture.format = .s16;
    device_config.capture.channels = 1;
    device_config.sampleRate = 44100;
    device_config.dataCallback = audioDataCallback;
    device_config.pUserData = null;

    audio_device = try audio_engine.?.createDevice(device_config);
    try audio_device.?.start();

    is_recording = true;
    recording_buffer.clearRetainingCapacity();
    
    std.debug.print("Recording started!\n", .{});
}

fn stopRecording() !void {
    std.debug.print("Stopping recording...\n", .{});
    
    if (audio_device) |device| {
        try device.stop();
        device.destroy();
        audio_device = null;
    }

    is_recording = false;
    try saveWavFile("recording.wav");
    
    std.debug.print("Recording stopped and saved to recording.wav\n", .{});
}

fn audioDataCallback(
    device: *zaudio.Device,
    output: ?*anyopaque,
    input: ?*const anyopaque,
    frame_count: u32,
) callconv(.C) void {
    _ = device;
    _ = output;
    
    if (input) |in_ptr| {
        const bytes_per_frame = 2;
        const byte_count = frame_count * bytes_per_frame;
        const input_bytes = @as([*]const u8, @ptrCast(in_ptr))[0..byte_count];
        
        recording_buffer.appendSlice(input_bytes) catch |err| {
            std.debug.print("Error appending audio data: {}\n", .{err});
        };
    }
}

fn saveWavFile(filename: []const u8) !void {
    const file = try std.fs.cwd().createFile(filename, .{});
    defer file.close();

    const sample_rate: u32 = 44100;
    const bits_per_sample: u16 = 16;
    const num_channels: u16 = 1;
    const byte_rate = sample_rate * num_channels * bits_per_sample / 8;
    const block_align = num_channels * bits_per_sample / 8;
    const data_size: u32 = @intCast(recording_buffer.items.len);
    const file_size: u32 = 36 + data_size;

    try file.writeAll("RIFF");
    try file.writeInt(u32, file_size, .little);
    try file.writeAll("WAVE");
    
    try file.writeAll("fmt ");
    try file.writeInt(u32, 16, .little);
    try file.writeInt(u16, 1, .little);
    try file.writeInt(u16, num_channels, .little);
    try file.writeInt(u32, sample_rate, .little);
    try file.writeInt(u32, byte_rate, .little);
    try file.writeInt(u16, block_align, .little);
    try file.writeInt(u16, bits_per_sample, .little);
    
    try file.writeAll("data");
    try file.writeInt(u32, data_size, .little);
    try file.writeAll(recording_buffer.items);
}
