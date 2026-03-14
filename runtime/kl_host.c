/*
 * kl_host.c — Host process for Klang DLL hot-reload.
 *
 * Architecture:
 *   - This host is compiled once and stays running.
 *   - Game code is compiled to a shared library (game.dll / game.so).
 *   - On each frame, the host calls game_update() and game_render() from the DLL.
 *   - When build/.reload appears, the host unloads the old DLL and loads the new one.
 *   - Game state (the void* instance) persists across reloads.
 *
 * Compile: gcc -o build/host runtime/kl_host.c -Iruntime -lm [raylib flags]
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* ---- platform-specific DLL loading ---- */

#ifdef _WIN32
  #define WIN32_LEAN_AND_MEAN
  #include <windows.h>

  typedef HMODULE DllHandle;

  static DllHandle dll_load(const char* path) {
      /* Copy to temp file so the original can be overwritten during rebuild */
      const char* tmp_path = "build\\game_live.dll";
      CopyFileA(path, tmp_path, FALSE);
      return LoadLibraryA(tmp_path);
  }

  static void dll_unload(DllHandle h) {
      if (h) FreeLibrary(h);
  }

  static void* dll_sym(DllHandle h, const char* name) {
      return (void*)GetProcAddress(h, name);
  }

  static int file_exists(const char* path) {
      DWORD attr = GetFileAttributesA(path);
      return (attr != INVALID_FILE_ATTRIBUTES);
  }

  static void file_delete(const char* path) {
      DeleteFileA(path);
  }

  static void sleep_ms(int ms) {
      Sleep(ms);
  }

#else
  #include <dlfcn.h>
  #include <unistd.h>
  #include <sys/stat.h>

  typedef void* DllHandle;

  static DllHandle dll_load(const char* path) {
      /* Copy to temp file so the original can be overwritten during rebuild */
      const char* tmp_path = "build/game_live.so";
      char cmd[512];
      snprintf(cmd, sizeof(cmd), "cp '%s' '%s'", path, tmp_path);
      system(cmd);
      return dlopen(tmp_path, RTLD_NOW);
  }

  static void dll_unload(DllHandle h) {
      if (h) dlclose(h);
  }

  static void* dll_sym(DllHandle h, const char* name) {
      return dlsym(h, name);
  }

  static int file_exists(const char* path) {
      struct stat st;
      return stat(path, &st) == 0;
  }

  static void file_delete(const char* path) {
      unlink(path);
  }

  static void sleep_ms(int ms) {
      usleep(ms * 1000);
  }
#endif

/* ---- function pointer types ---- */
typedef void* (*game_create_fn)(void);
typedef void  (*game_tick_fn)(void*);
typedef void  (*game_destroy_fn)(void*);

/* ---- main ---- */

#ifdef _WIN32
  #define DLL_PATH "build\\game.dll"
  #define RELOAD_SIGNAL "build\\.reload"
#else
  #define DLL_PATH "build/game.so"
  #define RELOAD_SIGNAL "build/.reload"
#endif

int main(int argc, char** argv) {
    (void)argc; (void)argv;

    /* Load game DLL */
    DllHandle dll = dll_load(DLL_PATH);
    if (!dll) {
        fprintf(stderr, "[host] failed to load game DLL: %s\n", DLL_PATH);
#ifndef _WIN32
        fprintf(stderr, "[host] dlerror: %s\n", dlerror());
#endif
        return 1;
    }

    game_create_fn  create_fn  = (game_create_fn)dll_sym(dll, "game_create");
    game_tick_fn    tick_fn    = (game_tick_fn)dll_sym(dll, "game_tick");
    game_destroy_fn destroy_fn = (game_destroy_fn)dll_sym(dll, "game_destroy");

    if (!create_fn) {
        fprintf(stderr, "[host] game DLL missing game_create()\n");
        dll_unload(dll);
        return 1;
    }

    /* Create game instance */
    void* game = create_fn();
    int generation = 1;

    fprintf(stderr, "[host] game loaded (gen %d)\n", generation);

    /* Main loop */
    while (1) {
        /* Check for reload signal */
        if (file_exists(RELOAD_SIGNAL)) {
            file_delete(RELOAD_SIGNAL);

            fprintf(stderr, "[host] reloading DLL...\n");

            /* Destroy old instance */
            if (destroy_fn && game) {
                destroy_fn(game);
                game = NULL;
            }

            /* Unload old DLL */
            dll_unload(dll);
            dll = NULL;

            /* Small delay to ensure file is fully written */
            sleep_ms(50);

            /* Load new DLL */
            dll = dll_load(DLL_PATH);
            if (!dll) {
                fprintf(stderr, "[host] failed to reload DLL, waiting...\n");
                sleep_ms(500);
                continue;
            }

            /* Re-resolve symbols */
            create_fn  = (game_create_fn)dll_sym(dll, "game_create");
            tick_fn    = (game_tick_fn)dll_sym(dll, "game_tick");
            destroy_fn = (game_destroy_fn)dll_sym(dll, "game_destroy");

            if (!create_fn) {
                fprintf(stderr, "[host] reloaded DLL missing game_create()\n");
                continue;
            }

            /* Recreate game instance */
            game = create_fn();
            generation++;
            fprintf(stderr, "[host] game reloaded (gen %d)\n", generation);
        }

        /* Call game_tick which invokes all non-main methods in order */
        if (tick_fn && game) tick_fn(game);

        /* If no tick function, this is a non-interactive program.
         * Run create once and exit. */
        if (!tick_fn) {
            break;
        }

        /* Throttle to ~60fps for non-raylib programs.
         * Raylib programs manage their own frame timing via set_target_fps(). */
        sleep_ms(16);
    }

    /* Cleanup */
    if (destroy_fn && game) destroy_fn(game);
    dll_unload(dll);

    return 0;
}
