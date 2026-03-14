/*
 * kl_host.c — Host process for Klang DLL hot-reload.
 *
 * Architecture:
 *   - This host is compiled once and stays running.
 *   - Game code is compiled to a shared library (game.dll / game.so).
 *   - game_create() allocates the instance, game_main() runs the full main().
 *   - game_main() runs in a separate thread so the host can watch for reloads.
 *   - On reload: host loads the NEW DLL alongside the old (never unloads old),
 *     calls game_patch() to update function pointers in the instance.
 *   - main() calls methods through instance-stored function pointers,
 *     so updated methods take effect on the next call.
 *   - Old DLLs stay loaded (main thread's stack references them).
 *
 * Compile: gcc -o build/host runtime/kl_host.c -Iruntime -lm [raylib flags]
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <signal.h>

static volatile int g_quit = 0;

static void signal_handler(int sig) {
    (void)sig;
    g_quit = 1;
}

/* ---- platform abstraction ---- */

#ifdef _WIN32
  #define WIN32_LEAN_AND_MEAN
  #include <windows.h>

  typedef HMODULE DllHandle;
  typedef HANDLE  ThreadHandle;

  static DllHandle dll_load(const char* path) {
      return LoadLibraryA(path);
  }
  static void* dll_sym(DllHandle h, const char* name) {
      return (void*)GetProcAddress(h, name);
  }
  static int file_exists(const char* path) {
      return GetFileAttributesA(path) != INVALID_FILE_ATTRIBUTES;
  }
  static void file_delete(const char* path) { DeleteFileA(path); }
  static void sleep_ms(int ms) { Sleep(ms); }

  /* Copy game.dll to a uniquely-named file and load that.
   * This keeps the original unlocked for the next rebuild. */
  static DllHandle load_dll_copy(const char* src, int gen) {
      char dest[256];
      snprintf(dest, sizeof(dest), "build\\_game_%d.dll", gen);
      CopyFileA(src, dest, FALSE);
      return dll_load(dest);
  }

  /* Thread wrapper */
  typedef DWORD WINAPI ThreadFunc(LPVOID);
  static ThreadHandle start_thread(ThreadFunc* fn, void* arg) {
      return CreateThread(NULL, 0, fn, arg, 0, NULL);
  }
  static int thread_alive(ThreadHandle h) {
      return WaitForSingleObject(h, 0) == WAIT_TIMEOUT;
  }
  static void thread_join(ThreadHandle h) {
      WaitForSingleObject(h, INFINITE);
      CloseHandle(h);
  }

#else
  #include <dlfcn.h>
  #include <unistd.h>
  #include <sys/stat.h>
  #include <pthread.h>

  typedef void* DllHandle;
  typedef pthread_t ThreadHandle;

  static DllHandle dll_load(const char* path) {
      return dlopen(path, RTLD_NOW);
  }
  static void* dll_sym(DllHandle h, const char* name) {
      return dlsym(h, name);
  }
  static int file_exists(const char* path) {
      struct stat st;
      return stat(path, &st) == 0;
  }
  static void file_delete(const char* path) { unlink(path); }
  static void sleep_ms(int ms) { usleep(ms * 1000); }

  static DllHandle load_dll_copy(const char* src, int gen) {
      char dest[256];
      snprintf(dest, sizeof(dest), "build/_game_%d.so", gen);
      char cmd[512];
      snprintf(cmd, sizeof(cmd), "cp '%s' '%s'", src, dest);
      system(cmd);
      return dll_load(dest);
  }

  static ThreadHandle start_thread(void* (*fn)(void*), void* arg) {
      pthread_t t;
      pthread_create(&t, NULL, fn, arg);
      return t;
  }
  static int thread_alive(ThreadHandle h) {
      return pthread_kill(h, 0) == 0;  /* 0 = thread exists */
  }
  static void thread_join(ThreadHandle h) { pthread_join(h, NULL); }
#endif

/* ---- DLL function signatures ---- */
typedef void* (*game_create_fn)(void);
typedef void  (*game_main_fn)(void*);
typedef void  (*game_patch_fn)(void*);
typedef void  (*game_destroy_fn)(void*);

/* ---- paths ---- */
#ifdef _WIN32
  #define DLL_PATH "build\\game.dll"
  #define RELOAD_SIGNAL "build\\.reload"
#else
  #define DLL_PATH "build/game.so"
  #define RELOAD_SIGNAL "build/.reload"
#endif

/* ---- main-thread runner ---- */
typedef struct {
    game_main_fn fn;
    void*        instance;
} MainThreadArgs;

#ifdef _WIN32
static DWORD WINAPI main_thread_func(LPVOID param) {
    MainThreadArgs* a = (MainThreadArgs*)param;
    a->fn(a->instance);
    free(a);
    return 0;
}
#else
static void* main_thread_func(void* param) {
    MainThreadArgs* a = (MainThreadArgs*)param;
    a->fn(a->instance);
    free(a);
    return NULL;
}
#endif

/* ---- reload watcher (runs on background thread) ---- */

typedef struct {
    void*            game;
    game_destroy_fn* destroy_fn_ptr;  /* points to main's destroy_fn */
    int              generation;
} WatcherArgs;

#ifdef _WIN32
static DWORD WINAPI watcher_thread_func(LPVOID param) {
#else
static void* watcher_thread_func(void* param) {
#endif
    WatcherArgs* w = (WatcherArgs*)param;
    int generation = w->generation;

    while (!g_quit) {
        if (file_exists(RELOAD_SIGNAL)) {
            file_delete(RELOAD_SIGNAL);
            generation++;

            DllHandle new_dll = load_dll_copy(DLL_PATH, generation);
            if (!new_dll) {
                fprintf(stderr, "[host] failed to load new DLL (gen %d)\n", generation);
            } else {
                game_patch_fn patch_fn = (game_patch_fn)dll_sym(new_dll, "game_patch");
                if (patch_fn) {
                    patch_fn(w->game);
                    fprintf(stderr, "[host] code patched (gen %d)\n", generation);
                }

                game_destroy_fn d = (game_destroy_fn)dll_sym(new_dll, "game_destroy");
                if (d) *(w->destroy_fn_ptr) = d;
            }
        }
        sleep_ms(100);
    }

    free(w);
#ifdef _WIN32
    return 0;
#else
    return NULL;
#endif
}

/* ---- entry point ---- */

int main(int argc, char** argv) {
    (void)argc; (void)argv;
    int generation = 0;

    signal(SIGINT, signal_handler);
    signal(SIGTERM, signal_handler);

    /* Clear stale reload signal */
    if (file_exists(RELOAD_SIGNAL)) file_delete(RELOAD_SIGNAL);

    /* Load initial DLL (copied so game.dll stays unlocked) */
    DllHandle dll_0 = load_dll_copy(DLL_PATH, generation);
    if (!dll_0) {
        fprintf(stderr, "[host] failed to load DLL\n");
        return 1;
    }

    game_create_fn  create_fn  = (game_create_fn)dll_sym(dll_0, "game_create");
    game_main_fn    main_fn    = (game_main_fn)dll_sym(dll_0, "game_main");
    game_destroy_fn destroy_fn = (game_destroy_fn)dll_sym(dll_0, "game_destroy");

    if (!create_fn || !main_fn) {
        fprintf(stderr, "[host] DLL missing game_create/game_main\n");
        return 1;
    }

    /* Create instance */
    void* game = create_fn();
    fprintf(stderr, "[host] game created (gen %d)\n", generation);

    /*
     * macOS requires all UI/window operations on the main thread (Cocoa).
     * So on macOS: run game_main() on main thread, watcher on background thread.
     * On other platforms: run game_main() on background thread, watch on main thread.
     */

    /* Start reload watcher on background thread */
    WatcherArgs* wargs = (WatcherArgs*)malloc(sizeof(WatcherArgs));
    wargs->game = game;
    wargs->destroy_fn_ptr = &destroy_fn;
    wargs->generation = generation;
    ThreadHandle watcher = start_thread(watcher_thread_func, wargs);

    /* Run game on main thread (required for macOS UI) */
    main_fn(game);

    /* game_main() returned — signal watcher to stop */
    g_quit = 1;
    thread_join(watcher);

    fprintf(stderr, "[host] main finished\n");
    if (destroy_fn && game) destroy_fn(game);

    /* Process exit cleans up all loaded DLLs */
    return 0;
}
