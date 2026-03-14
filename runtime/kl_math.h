#ifndef KL_MATH_H
#define KL_MATH_H

#include <math.h>
#include <float.h>

// ============================================================================
// Constants
// ============================================================================

#define KL_PI       3.14159265358979323846f
#define KL_TAU      6.28318530717958647692f
#define KL_E        2.71828182845904523536f
#define KL_DEG2RAD  (KL_PI / 180.0f)
#define KL_RAD2DEG  (180.0f / KL_PI)
#define KL_INF      HUGE_VALF
#define KL_EPSILON  FLT_EPSILON

// ============================================================================
// Utility functions (type-generic via _Generic)
// ============================================================================

static inline float kl_math_min_f(float a, float b) { return a < b ? a : b; }
static inline int   kl_math_min_i(int a, int b)     { return a < b ? a : b; }
static inline float kl_math_max_f(float a, float b) { return a > b ? a : b; }
static inline int   kl_math_max_i(int a, int b)     { return a > b ? a : b; }

#define kl_math_min(a, b) _Generic((a), float: kl_math_min_f, int: kl_math_min_i)((a), (b))
#define kl_math_max(a, b) _Generic((a), float: kl_math_max_f, int: kl_math_max_i)((a), (b))

static inline float kl_math_clamp_f(float x, float lo, float hi) {
    return kl_math_min_f(kl_math_max_f(x, lo), hi);
}
static inline int kl_math_clamp_i(int x, int lo, int hi) {
    return kl_math_min_i(kl_math_max_i(x, lo), hi);
}
#define kl_math_clamp(x, lo, hi) _Generic((x), \
    float: kl_math_clamp_f, \
    int: kl_math_clamp_i \
)((x), (lo), (hi))

static inline float kl_math_lerp(float a, float b, float t) {
    return a + (b - a) * t;
}

static inline float kl_math_sign_f(float x) { return (float)((x > 0.0f) - (x < 0.0f)); }
static inline int   kl_math_sign_i(int x)   { return (x > 0) - (x < 0); }
#define kl_math_sign(x) _Generic((x), float: kl_math_sign_f, int: kl_math_sign_i)(x)

static inline float kl_math_abs_f(float x) { return fabsf(x); }
static inline int   kl_math_abs_i(int x)   { return x < 0 ? -x : x; }
#define kl_math_abs(x) _Generic((x), float: kl_math_abs_f, int: kl_math_abs_i)(x)

// Degrees/Radians conversion
static inline float kl_math_deg2rad(float deg) { return deg * KL_DEG2RAD; }
static inline float kl_math_rad2deg(float rad) { return rad * KL_RAD2DEG; }

#endif // KL_MATH_H
