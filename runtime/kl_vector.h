#ifndef KL_VECTOR_H
#define KL_VECTOR_H

#include <math.h>

// ============================================================================
// Type definitions — when raylib is present, alias to raylib types directly
// so all raylib functions accept our types without casts
// ============================================================================

#ifdef KL_USE_RAYLIB
#include "raylib.h"

typedef Vector2 vec2;
typedef Vector3 vec3;
typedef Vector4 vec4;
typedef Vector4 quat;
// mat4 keeps its own definition (different struct layout from raylib Matrix)

#else

typedef struct { float x, y; } vec2;
typedef struct { float x, y, z; } vec3;
typedef struct { float x, y, z, w; } vec4;
typedef struct { float x, y, z, w; } quat;

#endif

// ============================================================================
// vec2
// ============================================================================

static inline vec2 vec2_new(float x, float y) { return (vec2){x, y}; }
static inline vec2 vec2_add(vec2 a, vec2 b) { return (vec2){a.x + b.x, a.y + b.y}; }
static inline vec2 vec2_sub(vec2 a, vec2 b) { return (vec2){a.x - b.x, a.y - b.y}; }
static inline vec2 vec2_scale(vec2 v, float s) { return (vec2){v.x * s, v.y * s}; }
static inline vec2 vec2_mul(vec2 a, vec2 b) { return (vec2){a.x * b.x, a.y * b.y}; }
static inline float vec2_dot(vec2 a, vec2 b) { return a.x * b.x + a.y * b.y; }
static inline float vec2_length(vec2 v) { return sqrtf(v.x * v.x + v.y * v.y); }
static inline vec2 vec2_normalize(vec2 v) {
    float len = vec2_length(v);
    if (len > 0.0f) { float inv = 1.0f / len; return (vec2){v.x * inv, v.y * inv}; }
    return v;
}
static inline float vec2_distance(vec2 a, vec2 b) { return vec2_length(vec2_sub(a, b)); }
static inline vec2 vec2_lerp(vec2 a, vec2 b, float t) {
    return (vec2){a.x + (b.x - a.x) * t, a.y + (b.y - a.y) * t};
}

// ============================================================================
// vec3
// ============================================================================

static inline vec3 vec3_new(float x, float y, float z) { return (vec3){x, y, z}; }
static inline vec3 vec3_zero(void)    { return (vec3){0, 0, 0}; }
static inline vec3 vec3_one(void)     { return (vec3){1, 1, 1}; }
static inline vec3 vec3_up(void)      { return (vec3){0, 1, 0}; }
static inline vec3 vec3_right(void)   { return (vec3){1, 0, 0}; }
static inline vec3 vec3_forward(void) { return (vec3){0, 0, -1}; }

static inline vec3 vec3_add(vec3 a, vec3 b) {
    return (vec3){a.x + b.x, a.y + b.y, a.z + b.z};
}
static inline vec3 vec3_sub(vec3 a, vec3 b) {
    return (vec3){a.x - b.x, a.y - b.y, a.z - b.z};
}
static inline vec3 vec3_scale(vec3 v, float s) {
    return (vec3){v.x * s, v.y * s, v.z * s};
}
static inline vec3 vec3_mul(vec3 a, vec3 b) {
    return (vec3){a.x * b.x, a.y * b.y, a.z * b.z};
}
static inline float vec3_dot(vec3 a, vec3 b) {
    return a.x * b.x + a.y * b.y + a.z * b.z;
}
static inline vec3 vec3_cross(vec3 a, vec3 b) {
    return (vec3){
        a.y * b.z - a.z * b.y,
        a.z * b.x - a.x * b.z,
        a.x * b.y - a.y * b.x
    };
}
static inline float vec3_length(vec3 v) {
    return sqrtf(v.x * v.x + v.y * v.y + v.z * v.z);
}
static inline float vec3_length_sq(vec3 v) {
    return v.x * v.x + v.y * v.y + v.z * v.z;
}
static inline vec3 vec3_normalize(vec3 v) {
    float len = vec3_length(v);
    if (len > 0.0f) {
        float inv = 1.0f / len;
        return (vec3){v.x * inv, v.y * inv, v.z * inv};
    }
    return v;
}
static inline float vec3_distance(vec3 a, vec3 b) {
    return vec3_length(vec3_sub(a, b));
}
static inline vec3 vec3_lerp(vec3 a, vec3 b, float t) {
    return (vec3){
        a.x + (b.x - a.x) * t,
        a.y + (b.y - a.y) * t,
        a.z + (b.z - a.z) * t
    };
}
static inline vec3 vec3_negate(vec3 v) {
    return (vec3){-v.x, -v.y, -v.z};
}
static inline vec3 vec3_reflect(vec3 v, vec3 normal) {
    float d = 2.0f * vec3_dot(v, normal);
    return vec3_sub(v, vec3_scale(normal, d));
}

// ============================================================================
// vec4
// ============================================================================

static inline vec4 vec4_new(float x, float y, float z, float w) { return (vec4){x, y, z, w}; }
static inline vec4 vec4_zero(void) { return (vec4){0, 0, 0, 0}; }
static inline vec4 vec4_one(void)  { return (vec4){1, 1, 1, 1}; }

static inline vec4 vec4_add(vec4 a, vec4 b) {
    return (vec4){a.x + b.x, a.y + b.y, a.z + b.z, a.w + b.w};
}
static inline vec4 vec4_sub(vec4 a, vec4 b) {
    return (vec4){a.x - b.x, a.y - b.y, a.z - b.z, a.w - b.w};
}
static inline vec4 vec4_scale(vec4 v, float s) {
    return (vec4){v.x * s, v.y * s, v.z * s, v.w * s};
}
static inline float vec4_dot(vec4 a, vec4 b) {
    return a.x * b.x + a.y * b.y + a.z * b.z + a.w * b.w;
}
static inline float vec4_length(vec4 v) {
    return sqrtf(v.x * v.x + v.y * v.y + v.z * v.z + v.w * v.w);
}
static inline vec4 vec4_normalize(vec4 v) {
    float len = vec4_length(v);
    if (len > 0.0f) {
        float inv = 1.0f / len;
        return (vec4){v.x * inv, v.y * inv, v.z * inv, v.w * inv};
    }
    return v;
}
static inline vec4 vec4_lerp(vec4 a, vec4 b, float t) {
    return (vec4){
        a.x + (b.x - a.x) * t,
        a.y + (b.y - a.y) * t,
        a.z + (b.z - a.z) * t,
        a.w + (b.w - a.w) * t
    };
}

// ============================================================================
// mat4 (4x4 matrix, column-major array — always our own type)
// ============================================================================

typedef struct {
    float m[16];
} mat4;

#define MAT4_AT(mat, row, col) ((mat).m[(col) * 4 + (row)])

static inline mat4 mat4_identity(void) {
    mat4 r = {{0}};
    r.m[0] = 1.0f; r.m[5] = 1.0f; r.m[10] = 1.0f; r.m[15] = 1.0f;
    return r;
}

static inline mat4 mat4_multiply(mat4 a, mat4 b) {
    mat4 r = {{0}};
    for (int col = 0; col < 4; col++) {
        for (int row = 0; row < 4; row++) {
            float sum = 0.0f;
            for (int k = 0; k < 4; k++) {
                sum += a.m[k * 4 + row] * b.m[col * 4 + k];
            }
            r.m[col * 4 + row] = sum;
        }
    }
    return r;
}

static inline mat4 mat4_translate(float x, float y, float z) {
    mat4 r = mat4_identity();
    r.m[12] = x; r.m[13] = y; r.m[14] = z;
    return r;
}

static inline mat4 mat4_scale_xyz(float x, float y, float z) {
    mat4 r = {{0}};
    r.m[0] = x; r.m[5] = y; r.m[10] = z; r.m[15] = 1.0f;
    return r;
}

static inline mat4 mat4_rotate(float angle_rad, float ax, float ay, float az) {
    float c = cosf(angle_rad);
    float s = sinf(angle_rad);
    float t = 1.0f - c;
    float len = sqrtf(ax * ax + ay * ay + az * az);
    if (len > 0.0f) { ax /= len; ay /= len; az /= len; }

    mat4 r = {{0}};
    r.m[0]  = t * ax * ax + c;
    r.m[1]  = t * ax * ay + s * az;
    r.m[2]  = t * ax * az - s * ay;
    r.m[4]  = t * ax * ay - s * az;
    r.m[5]  = t * ay * ay + c;
    r.m[6]  = t * ay * az + s * ax;
    r.m[8]  = t * ax * az + s * ay;
    r.m[9]  = t * ay * az - s * ax;
    r.m[10] = t * az * az + c;
    r.m[15] = 1.0f;
    return r;
}

static inline mat4 mat4_perspective(float fov_rad, float aspect, float near_p, float far_p) {
    float f = 1.0f / tanf(fov_rad / 2.0f);
    mat4 r = {{0}};
    r.m[0]  = f / aspect;
    r.m[5]  = f;
    r.m[10] = (far_p + near_p) / (near_p - far_p);
    r.m[11] = -1.0f;
    r.m[14] = (2.0f * far_p * near_p) / (near_p - far_p);
    return r;
}

static inline mat4 mat4_ortho(float left, float right, float bottom, float top, float near_p, float far_p) {
    mat4 r = {{0}};
    r.m[0]  = 2.0f / (right - left);
    r.m[5]  = 2.0f / (top - bottom);
    r.m[10] = -2.0f / (far_p - near_p);
    r.m[12] = -(right + left) / (right - left);
    r.m[13] = -(top + bottom) / (top - bottom);
    r.m[14] = -(far_p + near_p) / (far_p - near_p);
    r.m[15] = 1.0f;
    return r;
}

static inline mat4 mat4_look_at(vec3 eye, vec3 center, vec3 up) {
    vec3 f = vec3_normalize(vec3_sub(center, eye));
    vec3 s = vec3_normalize(vec3_cross(f, up));
    vec3 u = vec3_cross(s, f);

    mat4 r = mat4_identity();
    r.m[0] = s.x;  r.m[4] = s.y;  r.m[8]  = s.z;
    r.m[1] = u.x;  r.m[5] = u.y;  r.m[9]  = u.z;
    r.m[2] = -f.x; r.m[6] = -f.y; r.m[10] = -f.z;
    r.m[12] = -vec3_dot(s, eye);
    r.m[13] = -vec3_dot(u, eye);
    r.m[14] = vec3_dot(f, eye);
    return r;
}

static inline vec4 mat4_mul_vec4(mat4 m, vec4 v) {
    return (vec4){
        m.m[0] * v.x + m.m[4] * v.y + m.m[8]  * v.z + m.m[12] * v.w,
        m.m[1] * v.x + m.m[5] * v.y + m.m[9]  * v.z + m.m[13] * v.w,
        m.m[2] * v.x + m.m[6] * v.y + m.m[10] * v.z + m.m[14] * v.w,
        m.m[3] * v.x + m.m[7] * v.y + m.m[11] * v.z + m.m[15] * v.w
    };
}

// ============================================================================
// Quaternion operations (quat may be vec4 alias when raylib is present)
// ============================================================================

static inline quat quat_identity(void) { return (quat){0, 0, 0, 1}; }
static inline quat quat_new(float x, float y, float z, float w) { return (quat){x, y, z, w}; }

static inline quat quat_from_axis_angle(vec3 axis, float angle_rad) {
    float half = angle_rad * 0.5f;
    float s = sinf(half);
    vec3 n = vec3_normalize(axis);
    return (quat){n.x * s, n.y * s, n.z * s, cosf(half)};
}

static inline quat quat_multiply(quat a, quat b) {
    return (quat){
        a.w * b.x + a.x * b.w + a.y * b.z - a.z * b.y,
        a.w * b.y - a.x * b.z + a.y * b.w + a.z * b.x,
        a.w * b.z + a.x * b.y - a.y * b.x + a.z * b.w,
        a.w * b.w - a.x * b.x - a.y * b.y - a.z * b.z
    };
}

static inline float quat_length(quat q) {
    return sqrtf(q.x * q.x + q.y * q.y + q.z * q.z + q.w * q.w);
}

static inline quat quat_normalize(quat q) {
    float len = quat_length(q);
    if (len > 0.0f) {
        float inv = 1.0f / len;
        return (quat){q.x * inv, q.y * inv, q.z * inv, q.w * inv};
    }
    return q;
}

static inline quat quat_conjugate(quat q) {
    return (quat){-q.x, -q.y, -q.z, q.w};
}

static inline vec3 quat_rotate_vec3(quat q, vec3 v) {
    vec3 qv = {q.x, q.y, q.z};
    vec3 uv = vec3_cross(qv, v);
    vec3 uuv = vec3_cross(qv, uv);
    return vec3_add(v, vec3_add(vec3_scale(uv, 2.0f * q.w), vec3_scale(uuv, 2.0f)));
}

static inline quat quat_slerp(quat a, quat b, float t) {
    float dot = a.x * b.x + a.y * b.y + a.z * b.z + a.w * b.w;
    if (dot < 0.0f) {
        b = (quat){-b.x, -b.y, -b.z, -b.w};
        dot = -dot;
    }
    if (dot > 0.9995f) {
        quat r = {
            a.x + (b.x - a.x) * t, a.y + (b.y - a.y) * t,
            a.z + (b.z - a.z) * t, a.w + (b.w - a.w) * t
        };
        return quat_normalize(r);
    }
    float theta = acosf(dot);
    float sin_theta = sinf(theta);
    float wa = sinf((1.0f - t) * theta) / sin_theta;
    float wb = sinf(t * theta) / sin_theta;
    return (quat){
        wa * a.x + wb * b.x, wa * a.y + wb * b.y,
        wa * a.z + wb * b.z, wa * a.w + wb * b.w
    };
}

static inline mat4 quat_to_mat4(quat q) {
    float xx = q.x * q.x, yy = q.y * q.y, zz = q.z * q.z;
    float xy = q.x * q.y, xz = q.x * q.z, yz = q.y * q.z;
    float wx = q.w * q.x, wy = q.w * q.y, wz = q.w * q.z;

    mat4 r = {{0}};
    r.m[0]  = 1.0f - 2.0f * (yy + zz);
    r.m[1]  = 2.0f * (xy + wz);
    r.m[2]  = 2.0f * (xz - wy);
    r.m[4]  = 2.0f * (xy - wz);
    r.m[5]  = 1.0f - 2.0f * (xx + zz);
    r.m[6]  = 2.0f * (yz + wx);
    r.m[8]  = 2.0f * (xz + wy);
    r.m[9]  = 2.0f * (yz - wx);
    r.m[10] = 1.0f - 2.0f * (xx + yy);
    r.m[15] = 1.0f;
    return r;
}

#endif // KL_VECTOR_H
