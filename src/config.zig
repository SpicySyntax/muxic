const std = @import("std");

pub const Config = struct {
    default_device: ?[]const u8 = null,

    pub fn load(allocator: std.mem.Allocator) !Config {
        const file_path = "muxic_config.json";

        const file = std.fs.cwd().openFile(file_path, .{}) catch |err| {
            if (err == error.FileNotFound) {
                return Config{};
            }
            return err;
        };
        defer file.close();

        // Read the entire file
        const file_size = (try file.stat()).size;
        const buffer = try allocator.alloc(u8, file_size);
        defer allocator.free(buffer);

        const bytes_read = try file.read(buffer);
        if (bytes_read != file_size) return error.ReadError;

        // Parse JSON
        // Note: In a real app we might want to use a specific simplified parser or manage memory better
        // For now we duplicate the string to ensure it outlives the buffer
        const parsed = try std.json.parseFromSlice(Config, allocator, buffer, .{ .ignore_unknown_fields = true });
        defer parsed.deinit();

        var config = Config{};
        if (parsed.value.default_device) |dev| {
            config.default_device = try allocator.dupe(u8, dev);
        }

        return config;
    }

    pub fn save(self: Config) !void {
        const file_path = "muxic_config.json";

        const file = try std.fs.cwd().createFile(file_path, .{});
        defer file.close();

        var ws = std.json.writeStream(file.writer(), .{ .whitespace = .indent_2 });
        try ws.write(self);
    }

    pub fn deinit(self: *Config, allocator: std.mem.Allocator) void {
        if (self.default_device) |dev| {
            allocator.free(dev);
        }
    }
};
