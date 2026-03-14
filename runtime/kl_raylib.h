#ifndef KL_RAYLIB_H
#define KL_RAYLIB_H

// ============================================================================
// Klang Raylib Wrapper
// Provides the "rl" module for game development
// ============================================================================

#include "raylib.h"

// --- Color constructors ---
static inline Color kl_color(int r, int g, int b, int a) {
    return (Color){ (unsigned char)r, (unsigned char)g, (unsigned char)b, (unsigned char)a };
}

static inline Color kl_color_rgb(int r, int g, int b) {
    return (Color){ (unsigned char)r, (unsigned char)g, (unsigned char)b, 255 };
}

// --- Camera2D helper ---
static inline Camera2D kl_camera2d(vec2 offset, vec2 target, float rotation, float zoom) {
    Camera2D cam = {0};
    cam.offset = (Vector2){ offset.x, offset.y };
    cam.target = (Vector2){ target.x, target.y };
    cam.rotation = rotation;
    cam.zoom = zoom;
    return cam;
}

// --- Camera3D helper ---
static inline Camera3D kl_camera3d(vec3 position, vec3 target, vec3 up, float fovy, int projection) {
    Camera3D cam = {0};
    cam.position = (Vector3){ position.x, position.y, position.z };
    cam.target = (Vector3){ target.x, target.y, target.z };
    cam.up = (Vector3){ up.x, up.y, up.z };
    cam.fovy = fovy;
    cam.projection = projection;
    return cam;
}

// --- Draw helpers that accept klang types ---

// DrawLineV wrapper: takes vec2
static inline void kl_draw_line_v(vec2 start, vec2 end, Color color) {
    DrawLineV((Vector2){start.x, start.y}, (Vector2){end.x, end.y}, color);
}

// DrawCircleV wrapper: takes vec2
static inline void kl_draw_circle_v(vec2 center, float radius, Color color) {
    DrawCircleV((Vector2){center.x, center.y}, radius, color);
}

// DrawRectangleV wrapper: takes vec2 position and size
static inline void kl_draw_rectangle_v(vec2 position, vec2 size, Color color) {
    DrawRectangleV((Vector2){position.x, position.y}, (Vector2){size.x, size.y}, color);
}

// DrawTextureV wrapper: takes vec2 position
static inline void kl_draw_texture_v(Texture2D texture, vec2 position, Color tint) {
    DrawTextureV(texture, (Vector2){position.x, position.y}, tint);
}

// GetMousePosition wrapper: returns vec2
static inline vec2 kl_get_mouse_position(void) {
    Vector2 p = GetMousePosition();
    return (vec2){ p.x, p.y };
}

// GetScreenSize helper
static inline vec2 kl_get_screen_size(void) {
    return (vec2){ (float)GetScreenWidth(), (float)GetScreenHeight() };
}

// Rectangle helper
static inline Rectangle kl_rect(float x, float y, float w, float h) {
    return (Rectangle){ x, y, w, h };
}

// DrawTexturePro wrapper (common in 2D games)
static inline void kl_draw_texture_pro(Texture2D texture, Rectangle source, Rectangle dest, vec2 origin, float rotation, Color tint) {
    DrawTexturePro(texture, source, dest, (Vector2){origin.x, origin.y}, rotation, tint);
}

#endif // KL_RAYLIB_H
