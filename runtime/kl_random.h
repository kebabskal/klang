#ifndef KL_RANDOM_H
#define KL_RANDOM_H

#include <stdint.h>
#include <stdbool.h>

// ============================================================================
// Seeded PRNG — xoshiro128** (fast, good quality, 32-bit)
// ============================================================================

typedef struct {
    uint32_t s[4];
} KlRandom;

static inline uint32_t _kl_random_rotl(uint32_t x, int k) {
    return (x << k) | (x >> (32 - k));
}

static inline uint32_t _kl_random_next(KlRandom* r) {
    uint32_t result = _kl_random_rotl(r->s[1] * 5, 7) * 9;
    uint32_t t = r->s[1] << 9;
    r->s[2] ^= r->s[0];
    r->s[3] ^= r->s[1];
    r->s[1] ^= r->s[2];
    r->s[0] ^= r->s[3];
    r->s[2] ^= t;
    r->s[3] = _kl_random_rotl(r->s[3], 11);
    return result;
}

// Seed from a single uint32 using splitmix32 to fill state
static inline KlRandom kl_random_new(uint32_t seed) {
    KlRandom r;
    // splitmix32 to expand seed into 4 state words
    for (int i = 0; i < 4; i++) {
        seed += 0x9e3779b9u;
        uint32_t z = seed;
        z ^= z >> 16; z *= 0x85ebca6bu;
        z ^= z >> 13; z *= 0xc2b2ae35u;
        z ^= z >> 16;
        r.s[i] = z;
    }
    return r;
}

// Random int in [min, max] (inclusive)
static inline int kl_random_rangei(KlRandom* r, int min, int max) {
    if (min >= max) return min;
    uint32_t range = (uint32_t)(max - min + 1);
    return min + (int)(_kl_random_next(r) % range);
}

// Random float in [0.0, 1.0)
static inline float kl_random_float(KlRandom* r) {
    return (_kl_random_next(r) >> 8) * (1.0f / 16777216.0f);
}

// Random float in [min, max)
static inline float kl_random_rangef(KlRandom* r, float min, float max) {
    return min + kl_random_float(r) * (max - min);
}

// Random bool
static inline bool kl_random_bool(KlRandom* r) {
    return (_kl_random_next(r) & 1) != 0;
}

// Random int in [0, n) — alias for convenience
static inline int kl_random_int(KlRandom* r, int n) {
    if (n <= 0) return 0;
    return (int)(_kl_random_next(r) % (uint32_t)n);
}

#endif // KL_RANDOM_H
