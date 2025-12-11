const std = @import("std");

pub fn main() void {
    const json = std.json;
    inline for (@typeInfo(json).Struct.decls) |decl| {
        std.debug.print("{s}\n", .{decl.name});
    }
}
