#ifndef KL_RUNTIME_H
#define KL_RUNTIME_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <stdint.h>

// Standard library headers (kl_vector.h may include raylib.h, which must come before windows.h)
#include "kl_math.h"
#include "kl_vector.h"
#include "kl_random.h"

#ifdef _WIN32
  #define WIN32_LEAN_AND_MEAN
  #ifdef KL_USE_RAYLIB
    // Avoid Windows API name conflicts with raylib
    #define NOGDI
    #define NOUSER
  #endif
  #include <windows.h>
  #ifdef KL_USE_RAYLIB
    #undef near
    #undef far
  #endif
#else
  #include <unistd.h>
#endif

// ============================================================================
// Memory Management — Refcounted objects with auto-weak cycle prevention
// ============================================================================

// --- Forward declarations ---

typedef struct KlObject KlObject;
typedef struct KlWeakSlot KlWeakSlot;
typedef void (*KlDestructor)(KlObject*);
typedef void (*KlTracer)(KlObject*, void (*visit)(KlObject*));

// --- Object header (embedded at start of every heap-allocated object) ---

typedef struct {
    int type_id;
    int refcount;
    KlDestructor destructor;  // per-class: releases strong fields
    KlTracer tracer;          // per-class: visits strong refs (for cycle collector)
    KlObject* gc_next;        // intrusive linked list for cycle collector
    KlWeakSlot* weak_slot;    // lazily allocated, shared by all weak refs to this object
    int gc_color;             // used by cycle collector (index during trial deletion)
} KlHeader;

struct KlObject {
    KlHeader _header;
};

// --- Weak reference slot ---
// When someone takes a weak ref, a slot is created (or reused).
// All weak refs to the same object share one slot.
// When the object dies, slot->target is set to NULL.

struct KlWeakSlot {
    KlObject* target;   // points to object, or NULL if dead
    int ref_count;      // number of weak field holders pointing to this slot
};

// --- GC globals ---

static KlObject* kl_gc_head = NULL;
static int kl_gc_alloc_count = 0;
#define KL_GC_THRESHOLD 1024

// Forward declare cycle collector
static void kl_gc_collect_cycles(void);

// --- Raw allocator (for non-RC internal use: list backing arrays, etc.) ---

static inline void* kl_alloc(size_t size) {
    void* ptr = calloc(1, size);
    if (!ptr) {
        fprintf(stderr, "klang: out of memory\n");
        exit(1);
    }
    return ptr;
}

// --- GC linked list management ---

static inline void kl_gc_unlink(KlObject* obj) {
    if (kl_gc_head == obj) {
        kl_gc_head = obj->_header.gc_next;
        return;
    }
    KlObject* prev = kl_gc_head;
    while (prev && prev->_header.gc_next != obj) {
        prev = prev->_header.gc_next;
    }
    if (prev) {
        prev->_header.gc_next = obj->_header.gc_next;
    }
}

// --- Weak ref helpers ---

static inline KlWeakSlot* kl_weak_slot_get(KlObject* obj) {
    if (!obj) return NULL;
    if (!obj->_header.weak_slot) {
        obj->_header.weak_slot = (KlWeakSlot*)calloc(1, sizeof(KlWeakSlot));
        obj->_header.weak_slot->target = obj;
        obj->_header.weak_slot->ref_count = 0;
    }
    return obj->_header.weak_slot;
}

static inline void* kl_weak_read(KlWeakSlot* slot) {
    if (!slot) return NULL;
    return slot->target;
}

static inline void kl_weak_assign(KlWeakSlot** field, void* new_val) {
    // Release old slot
    if (*field) {
        (*field)->ref_count--;
        if ((*field)->target == NULL && (*field)->ref_count <= 0) {
            free(*field);
        }
    }
    // Assign new slot
    if (new_val) {
        KlWeakSlot* slot = kl_weak_slot_get((KlObject*)new_val);
        slot->ref_count++;
        *field = slot;
    } else {
        *field = NULL;
    }
}

static inline void kl_weak_release(KlWeakSlot** field) {
    kl_weak_assign(field, NULL);
}

// --- Core retain / release ---

static inline void kl_retain(void* ptr) {
    if (!ptr) return;
    ((KlObject*)ptr)->_header.refcount++;
}

static inline void kl_release(void* ptr) {
    if (!ptr) return;
    KlObject* obj = (KlObject*)ptr;
    if (--obj->_header.refcount <= 0) {
        // Invalidate weak slot
        if (obj->_header.weak_slot) {
            obj->_header.weak_slot->target = NULL;
            if (obj->_header.weak_slot->ref_count <= 0) {
                free(obj->_header.weak_slot);
            }
            obj->_header.weak_slot = NULL;
        }
        // Call class-specific destructor (releases strong fields)
        if (obj->_header.destructor) {
            obj->_header.destructor(obj);
        }
        // Unlink from GC list
        kl_gc_unlink(obj);
        free(obj);
    }
}

// --- RC allocator ---

static inline void* kl_alloc_rc(size_t size, KlDestructor dtor) {
    KlObject* obj = (KlObject*)calloc(1, size);
    if (!obj) {
        fprintf(stderr, "klang: out of memory\n");
        exit(1);
    }
    obj->_header.refcount = 1;
    obj->_header.destructor = dtor;
    obj->_header.tracer = NULL;
    obj->_header.gc_next = kl_gc_head;
    obj->_header.weak_slot = NULL;
    kl_gc_head = obj;
    kl_gc_alloc_count++;
    if (kl_gc_alloc_count >= KL_GC_THRESHOLD) {
        kl_gc_collect_cycles();
        kl_gc_alloc_count = 0;
    }
    return obj;
}

// --- Strong field assignment: retain new, release old, then assign ---

static inline void kl_strong_assign(void** field, void* new_val) {
    if (*field == new_val) return;
    kl_retain(new_val);
    kl_release(*field);
    *field = new_val;
}

// --- Cycle collector (Bacon's trial deletion) ---

// Temp refcount storage for trial deletion
static int* kl_gc_trial_counts = NULL;
static int kl_gc_trial_cap = 0;
static int kl_gc_trial_len = 0;

// Assign each object an index for trial deletion
static void kl_gc_trial_dec_visitor(KlObject* obj) {
    // Decrement trial refcount of visited object
    // We use gc_color to store the index into trial_counts
    if (obj && obj->_header.gc_color >= 0 && obj->_header.gc_color < kl_gc_trial_len) {
        kl_gc_trial_counts[obj->_header.gc_color]--;
    }
}

static void kl_gc_restore_visitor(KlObject* obj) {
    if (!obj) return;
    int idx = obj->_header.gc_color;
    if (idx >= 0 && idx < kl_gc_trial_len && kl_gc_trial_counts[idx] <= 0) {
        kl_gc_trial_counts[idx] = obj->_header.refcount;
        // Recursively restore children
        if (obj->_header.tracer) {
            obj->_header.tracer(obj, kl_gc_restore_visitor);
        }
    }
}

static void kl_gc_collect_cycles(void) {
    // Count objects
    int count = 0;
    KlObject* obj = kl_gc_head;
    while (obj) { count++; obj = obj->_header.gc_next; }
    if (count == 0) return;

    // Allocate trial refcounts
    if (count > kl_gc_trial_cap) {
        kl_gc_trial_cap = count * 2;
        kl_gc_trial_counts = (int*)realloc(kl_gc_trial_counts, sizeof(int) * kl_gc_trial_cap);
    }
    kl_gc_trial_len = count;

    // Phase 1: Assign indices and copy refcounts
    int idx = 0;
    obj = kl_gc_head;
    while (obj) {
        obj->_header.gc_color = idx;
        kl_gc_trial_counts[idx] = obj->_header.refcount;
        idx++;
        obj = obj->_header.gc_next;
    }

    // Phase 2: Trial-decrement — for each object, decrement its children's trial counts
    obj = kl_gc_head;
    while (obj) {
        if (obj->_header.tracer) {
            obj->_header.tracer(obj, kl_gc_trial_dec_visitor);
        }
        obj = obj->_header.gc_next;
    }

    // Phase 3: Objects with trial count > 0 are roots. Restore their subgraphs.
    obj = kl_gc_head;
    while (obj) {
        int i = obj->_header.gc_color;
        if (kl_gc_trial_counts[i] > 0) {
            // This is a root — restore all its children
            if (obj->_header.tracer) {
                obj->_header.tracer(obj, kl_gc_restore_visitor);
            }
        }
        obj = obj->_header.gc_next;
    }

    // Phase 4: Sweep — objects with trial count <= 0 are cyclic garbage
    KlObject* prev = NULL;
    obj = kl_gc_head;
    while (obj) {
        KlObject* next = obj->_header.gc_next;
        int i = obj->_header.gc_color;
        if (kl_gc_trial_counts[i] <= 0) {
            // Cyclic garbage — invalidate weak slot
            if (obj->_header.weak_slot) {
                obj->_header.weak_slot->target = NULL;
                if (obj->_header.weak_slot->ref_count <= 0) {
                    free(obj->_header.weak_slot);
                }
            }
            // Call destructor (but don't recurse into release for cycle members)
            if (obj->_header.destructor) {
                obj->_header.destructor(obj);
            }
            // Unlink
            if (prev) prev->_header.gc_next = next;
            else kl_gc_head = next;
            free(obj);
        } else {
            prev = obj;
        }
        obj = next;
    }
}

// ============================================================================
// Print
// ============================================================================

static inline void kl_print_str(const char* s) {
    printf("%s\n", s); fflush(stdout);
}

static inline void kl_print_int(int v) {
    printf("%d\n", v); fflush(stdout);
}

static inline void kl_print_float(float v) {
    printf("%g\n", v); fflush(stdout);
}

static inline void kl_print_bool(bool v) {
    printf("%s\n", v ? "true" : "false"); fflush(stdout);
}

#define print(x) _Generic((x), \
    const char*: kl_print_str, \
    char*: kl_print_str, \
    int: kl_print_int, \
    float: kl_print_float, \
    _Bool: kl_print_bool \
)(x)

// Inline print (no newline) for multi-arg print
static inline void kl_print_inline_str(const char* s) { printf("%s", s); }
static inline void kl_print_inline_int(int v) { printf("%d", v); }
static inline void kl_print_inline_float(float v) { printf("%g", v); }
static inline void kl_print_inline_bool(bool v) { printf("%s", v ? "true" : "false"); }
static inline void kl_print_nl(void) { printf("\n"); fflush(stdout); }

#define kl_print_inline(x) _Generic((x), \
    const char*: kl_print_inline_str, \
    char*: kl_print_inline_str, \
    int: kl_print_inline_int, \
    float: kl_print_inline_float, \
    _Bool: kl_print_inline_bool \
)(x)

// ============================================================================
// Wait / Sleep
// ============================================================================

#ifdef _WIN32
static inline void kl_wait(float seconds) {
    Sleep((DWORD)(seconds * 1000.0f));
}
#else
static inline void kl_wait(float seconds) {
    usleep((useconds_t)(seconds * 1000000.0f));
}
#endif

// ============================================================================
// Closures
// ============================================================================

typedef struct {
    KlHeader _header;
    void* fn;
    void* captures;
    KlDestructor captures_dtor;
} KlClosure;

static void kl_closure_destroy(KlObject* obj) {
    KlClosure* cl = (KlClosure*)obj;
    if (cl->captures && cl->captures_dtor) {
        cl->captures_dtor((KlObject*)cl->captures);
    }
    free(cl->captures);
}

// ============================================================================
// Dynamic list
// ============================================================================

typedef struct {
    KlHeader _header;
    void** data;
    int count;
    int capacity;
    bool items_are_rc;
} KlList;

// Forward declare
static void kl_list_destroy(KlObject* obj);
static void kl_list_trace(KlObject* obj, void (*visit)(KlObject*));

static inline KlList* kl_list_new(bool items_rc) {
    KlList* list = (KlList*)kl_alloc_rc(sizeof(KlList), kl_list_destroy);
    list->_header.tracer = NULL; // list tracer set below if items are RC
    list->capacity = 8;
    list->data = (void**)kl_alloc(sizeof(void*) * list->capacity);
    list->items_are_rc = items_rc;
    if (items_rc) list->_header.tracer = kl_list_trace;
    return list;
}

static void kl_list_trace(KlObject* obj, void (*visit)(KlObject*)) {
    KlList* list = (KlList*)obj;
    if (list->items_are_rc) {
        for (int i = 0; i < list->count; i++) {
            if (list->data[i]) visit((KlObject*)list->data[i]);
        }
    }
}

static void kl_list_destroy(KlObject* obj) {
    KlList* list = (KlList*)obj;
    if (list->items_are_rc) {
        for (int i = 0; i < list->count; i++) {
            kl_release(list->data[i]);
        }
    }
    free(list->data);
}

static inline void kl_list_push(KlList* list, void* item) {
    if (list->items_are_rc) kl_retain(item);
    if (list->count >= list->capacity) {
        list->capacity *= 2;
        list->data = (void**)realloc(list->data, sizeof(void*) * list->capacity);
    }
    list->data[list->count++] = item;
}

static inline void* kl_list_get(KlList* list, int index) {
    if (index < 0 || index >= list->count) return NULL;
    return list->data[index];
}

static inline void kl_list_set(KlList* list, int index, void* item) {
    if (index < 0 || index >= list->count) return;
    if (list->items_are_rc) {
        kl_retain(item);
        if (list->data[index]) kl_release(list->data[index]);
    }
    list->data[index] = item;
}

static inline void kl_list_remove(KlList* list, int index) {
    if (index < 0 || index >= list->count) return;
    if (list->items_are_rc && list->data[index]) {
        kl_release(list->data[index]);
    }
    for (int i = index; i < list->count - 1; i++) {
        list->data[i] = list->data[i + 1];
    }
    list->count--;
}

static inline void kl_list_insert(KlList* list, int index, void* item) {
    if (index < 0) index = 0;
    if (index > list->count) index = list->count;
    if (list->items_are_rc) kl_retain(item);
    if (list->count >= list->capacity) {
        list->capacity *= 2;
        list->data = (void**)realloc(list->data, sizeof(void*) * list->capacity);
    }
    for (int i = list->count; i > index; i--) {
        list->data[i] = list->data[i - 1];
    }
    list->data[index] = item;
    list->count++;
}

static inline void* kl_list_pop(KlList* list) {
    if (list->count == 0) return NULL;
    list->count--;
    void* item = list->data[list->count];
    return item;
}

static inline void* kl_list_first(KlList* list) {
    if (list->count == 0) return NULL;
    return list->data[0];
}

static inline void* kl_list_last(KlList* list) {
    if (list->count == 0) return NULL;
    return list->data[list->count - 1];
}

static inline void kl_list_clear(KlList* list) {
    if (list->items_are_rc) {
        for (int i = 0; i < list->count; i++) {
            if (list->data[i]) kl_release(list->data[i]);
        }
    }
    list->count = 0;
}

static inline void kl_list_reverse(KlList* list) {
    for (int i = 0, j = list->count - 1; i < j; i++, j--) {
        void* tmp = list->data[i];
        list->data[i] = list->data[j];
        list->data[j] = tmp;
    }
}

static inline KlList* kl_list_clone(KlList* list) {
    KlList* result = kl_list_new(list->items_are_rc);
    for (int i = 0; i < list->count; i++) {
        kl_list_push(result, list->data[i]);
    }
    return result;
}

static inline KlList* kl_list_slice(KlList* list, int start, int end) {
    if (start < 0) start = 0;
    if (end > list->count) end = list->count;
    if (start >= end) return kl_list_new(list->items_are_rc);
    KlList* result = kl_list_new(list->items_are_rc);
    for (int i = start; i < end; i++) {
        kl_list_push(result, list->data[i]);
    }
    return result;
}

static inline bool kl_list_contains(KlList* list, void* item) {
    for (int i = 0; i < list->count; i++) {
        if (list->data[i] == item) return true;
    }
    return false;
}

static inline int kl_list_index_of(KlList* list, void* item) {
    for (int i = 0; i < list->count; i++) {
        if (list->data[i] == item) return i;
    }
    return -1;
}

// IO (included after KlList is defined)
#include "kl_io.h"

// Raylib wrapper (conditional)
#ifdef KL_USE_RAYLIB
#include "kl_raylib.h"
#endif

#endif // KL_RUNTIME_H
