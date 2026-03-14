#ifndef KL_IO_H
#define KL_IO_H

#include <stdio.h>
#include <sys/stat.h>

#ifdef _WIN32
#include <direct.h>
#include <io.h>
// Prevent windows.h from defining symbols that conflict with raylib
#ifndef NOGDI
#define NOGDI
#define _KL_UNDEF_NOGDI
#endif
#ifndef NOUSER
#define NOUSER
#define _KL_UNDEF_NOUSER
#endif
#ifndef WIN32_LEAN_AND_MEAN
#define WIN32_LEAN_AND_MEAN
#define _KL_UNDEF_LEAN
#endif
#include <windows.h>
#ifdef _KL_UNDEF_NOGDI
#undef NOGDI
#undef _KL_UNDEF_NOGDI
#endif
#ifdef _KL_UNDEF_NOUSER
#undef NOUSER
#undef _KL_UNDEF_NOUSER
#endif
#ifdef _KL_UNDEF_LEAN
#undef WIN32_LEAN_AND_MEAN
#undef _KL_UNDEF_LEAN
#endif
#define KL_MKDIR(path) _mkdir(path)
#define KL_ACCESS(path, mode) _access(path, mode)
#else
#include <unistd.h>
#include <dirent.h>
#define KL_MKDIR(path) mkdir(path, 0755)
#define KL_ACCESS(path, mode) access(path, mode)
#endif

// ============================================================================
// File operations
// ============================================================================

// Read entire file into a heap-allocated string
static inline const char* kl_io_read_file(const char* path) {
    FILE* f = fopen(path, "rb");
    if (!f) return "";
    fseek(f, 0, SEEK_END);
    long size = ftell(f);
    fseek(f, 0, SEEK_SET);
    char* buf = (char*)malloc(size + 1);
    if (!buf) { fclose(f); return ""; }
    fread(buf, 1, size, f);
    buf[size] = '\0';
    fclose(f);
    return buf;
}

// Write string to file (creates or overwrites)
static inline bool kl_io_write_file(const char* path, const char* content) {
    FILE* f = fopen(path, "wb");
    if (!f) return false;
    size_t len = strlen(content);
    size_t written = fwrite(content, 1, len, f);
    fclose(f);
    return written == len;
}

// Append string to file
static inline bool kl_io_append_file(const char* path, const char* content) {
    FILE* f = fopen(path, "ab");
    if (!f) return false;
    size_t len = strlen(content);
    size_t written = fwrite(content, 1, len, f);
    fclose(f);
    return written == len;
}

// Check if file exists
static inline bool kl_io_file_exists(const char* path) {
    return KL_ACCESS(path, 0) == 0;
}

// Delete a file
static inline bool kl_io_delete_file(const char* path) {
    return remove(path) == 0;
}

// ============================================================================
// Directory operations
// ============================================================================

// Create a directory
static inline bool kl_io_create_dir(const char* path) {
    return KL_MKDIR(path) == 0;
}

// Check if directory exists
static inline bool kl_io_dir_exists(const char* path) {
    struct stat st;
    if (stat(path, &st) != 0) return false;
    return (st.st_mode & S_IFDIR) != 0;
}

// List directory entries (returns KlList* of const char*)
static inline KlList* kl_io_list_dir(const char* path) {
    KlList* list = kl_list_new(false);
#ifdef _WIN32
    char search_path[MAX_PATH];
    snprintf(search_path, MAX_PATH, "%s\\*", path);
    WIN32_FIND_DATAA fd;
    HANDLE h = FindFirstFileA(search_path, &fd);
    if (h == INVALID_HANDLE_VALUE) return list;
    do {
        if (fd.cFileName[0] == '.' && (fd.cFileName[1] == '\0' ||
            (fd.cFileName[1] == '.' && fd.cFileName[2] == '\0'))) continue;
        char* name = (char*)malloc(strlen(fd.cFileName) + 1);
        strcpy(name, fd.cFileName);
        kl_list_push(list, (void*)name);
    } while (FindNextFileA(h, &fd));
    FindClose(h);
#else
    DIR* dir = opendir(path);
    if (!dir) return list;
    struct dirent* entry;
    while ((entry = readdir(dir)) != NULL) {
        if (entry->d_name[0] == '.' && (entry->d_name[1] == '\0' ||
            (entry->d_name[1] == '.' && entry->d_name[2] == '\0'))) continue;
        char* name = (char*)malloc(strlen(entry->d_name) + 1);
        strcpy(name, entry->d_name);
        kl_list_push(list, (void*)name);
    }
    closedir(dir);
#endif
    return list;
}

#endif // KL_IO_H
