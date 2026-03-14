package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/klang-lang/klang/internal/analysis"
	"github.com/klang-lang/klang/internal/codegen"
	"github.com/klang-lang/klang/internal/errs"
	"github.com/klang-lang/klang/internal/lexer"
	"github.com/klang-lang/klang/internal/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, "missing command"))
		fmt.Fprintln(os.Stderr, "usage: kl <command> [args]")
		fmt.Fprintln(os.Stderr, "commands: build, run, dev")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		cmdBuild()
	case "run":
		cmdRun()
	case "dev":
		cmdDev()
	case "lsp":
		cmdLsp()
	default:
		fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, fmt.Sprintf("unknown command '%s'", os.Args[1])))
		fmt.Fprintln(os.Stderr, "commands: build, run, dev, lsp")
		os.Exit(1)
	}
}

// ---------- build helpers (shared by build, run, dev) ----------

type buildResult struct {
	outputPath     string
	ok             bool
	hasHotReload   bool // true if the program has update()/render() methods
}

// doBuild compiles .k files and links them. Returns the output binary path on success.
// dllMode=true compiles to a shared library instead of an executable.
func doBuild(files []string, mode string, dllMode bool) buildResult {
	if len(files) == 0 {
		entries, err := os.ReadDir(".")
		if err != nil {
			fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, fmt.Sprintf("cannot read directory: %v", err)))
			return buildResult{}
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".k") {
				files = append(files, e.Name())
			}
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, "no .k files found"))
		return buildResult{}
	}

	os.MkdirAll("build", 0755)

	var cFiles []string
	needsRaylib := false
	hasHotReload := false
	for _, file := range files {
		cFile, usesRL, hotReload, err := compileFile(file, dllMode)
		if err != nil {
			return buildResult{}
		}
		cFiles = append(cFiles, cFile)
		if usesRL {
			needsRaylib = true
		}
		if hotReload {
			hasHotReload = true
		}
	}

	runtimeDir := findRuntime()

	// Determine output path
	var outputName string
	if dllMode {
		if isWindows() {
			outputName = "build\\game.dll"
		} else {
			outputName = "build/game.so"
		}
	} else {
		if isWindows() {
			outputName = "build\\game.exe"
		} else {
			outputName = "build/game"
		}
	}

	args := []string{"-o", outputName}

	if dllMode {
		args = append(args, "-shared")
		if !isWindows() {
			args = append(args, "-fPIC")
		}
	}

	if mode == "release" {
		args = append(args, "-O2", "-DNDEBUG")
	} else {
		args = append(args, "-g", "-O0")
	}
	args = append(args, fmt.Sprintf("-I%s", runtimeDir))

	if needsRaylib {
		raylibDir := findRaylib()
		if raylibDir != "" {
			args = append(args, fmt.Sprintf("-I%s/include", raylibDir))
			args = append(args, fmt.Sprintf("-L%s/lib", raylibDir))
		}
	}

	args = append(args, cFiles...)
	args = append(args, "-lm")

	if needsRaylib {
		if dllMode && isWindows() {
			// DLL mode: link against host import lib (raylib lives in the host)
			args = append(args, fmt.Sprintf("-Lbuild"))
			args = append(args, "-lhost")
		} else if isWindows() {
			args = append(args, "-lraylib", "-lopengl32", "-lgdi32", "-lwinmm")
		} else {
			args = append(args, "-lraylib", "-lGL", "-lpthread", "-ldl", "-lrt", "-lX11")
		}
	}

	cc := findCC()
	label := "exe"
	if dllMode {
		label = "dll"
	}
	fmt.Printf("compiling %s with %s (%s mode)...\n", label, cc, mode)
	if needsRaylib {
		fmt.Println("  (with raylib)")
	}

	cmd := exec.Command(cc, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, "C compilation failed"))
		return buildResult{}
	}

	fmt.Printf("built: %s\n", outputName)
	return buildResult{outputPath: outputName, ok: true, hasHotReload: hasHotReload}
}

// buildHost compiles the host executable (kl_host.c) that loads the game DLL.
// Only needs to be built once per dev session.
func buildHost(needsRaylib bool) buildResult {
	runtimeDir := findRuntime()
	hostSrc := filepath.Join(runtimeDir, "kl_host.c")

	if _, err := os.Stat(hostSrc); err != nil {
		fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, fmt.Sprintf("host runtime not found: %s", hostSrc)))
		return buildResult{}
	}

	var outputName string
	if isWindows() {
		outputName = "build\\host.exe"
	} else {
		outputName = "build/host"
	}

	cc := findCC()
	args := []string{"-o", outputName, "-g", "-O0", fmt.Sprintf("-I%s", runtimeDir), hostSrc, "-lm"}

	if needsRaylib {
		raylibDir := findRaylib()
		if raylibDir != "" {
			args = append(args, fmt.Sprintf("-I%s/include", raylibDir))
			args = append(args, fmt.Sprintf("-L%s/lib", raylibDir))
		}
		if isWindows() {
			// Export all raylib symbols so the DLL can import them from the host.
			// --whole-archive forces all raylib objects into the host (even if host doesn't call them).
			args = append(args, "-Wl,--export-all-symbols")
			args = append(args, fmt.Sprintf("-Wl,--out-implib,build%chost.lib", os.PathSeparator))
			args = append(args, "-Wl,--whole-archive", "-lraylib", "-Wl,--no-whole-archive")
			args = append(args, "-lopengl32", "-lgdi32", "-lwinmm")
		} else {
			args = append(args, "-lraylib", "-lGL", "-lpthread", "-ldl", "-lrt", "-lX11")
		}
	}

	fmt.Printf("compiling host with %s...\n", cc)
	cmd := exec.Command(cc, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, "host compilation failed"))
		return buildResult{}
	}

	fmt.Printf("built: %s\n", outputName)
	return buildResult{outputPath: outputName, ok: true}
}

func parseArgs() (files []string, mode string) {
	mode = "debug"
	for _, arg := range os.Args[2:] {
		switch arg {
		case "release":
			mode = "release"
		case "debug":
			mode = "debug"
		default:
			files = append(files, arg)
		}
	}
	return
}

// ---------- commands ----------

func cmdLsp() {
	// Find the kl-lsp binary next to kl, or in build/ directory
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	candidates := []string{
		filepath.Join(dir, "kl-lsp.exe"),
		filepath.Join(dir, "kl-lsp"),
		filepath.Join("build", "kl-lsp.exe"),
		filepath.Join("build", "kl-lsp"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			cmd := exec.Command(c)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				os.Exit(1)
			}
			return
		}
	}
	fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, "kl-lsp binary not found"))
	fmt.Fprintln(os.Stderr, "build it with: go build -o build/kl-lsp ./cmd/kl-lsp")
	os.Exit(1)
}

func cmdBuild() {
	files, mode := parseArgs()
	res := doBuild(files, mode, false)
	if !res.ok {
		os.Exit(1)
	}
}

func cmdRun() {
	files, mode := parseArgs()
	res := doBuild(files, mode, false)
	if !res.ok {
		os.Exit(1)
	}

	fmt.Println("running...")
	cmd := exec.Command(res.outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", errs.FormatSimple(errs.Error, fmt.Sprintf("process exited with %v", err)))
		os.Exit(1)
	}
}

// checkHotReload does a quick lex+parse of the source files to detect
// if the main class has non-main methods (making it eligible for hot reload).
func checkHotReload(files []string) bool {
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lex := lexer.New(src)
		tokens := lex.Tokenize()
		p := parser.New(tokens)
		file, err := p.Parse()
		if err != nil {
			continue
		}
		gen := codegen.New(file)
		if gen.HasHotReloadMethods() {
			return true
		}
	}
	return false
}

func cmdDev() {
	files, _ := parseArgs()

	// Resolve which .k files to watch
	watchFiles := files
	if len(watchFiles) == 0 {
		entries, err := os.ReadDir(".")
		if err != nil {
			fatal("cannot read directory: %v", err)
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".k") {
				watchFiles = append(watchFiles, e.Name())
			}
		}
	}
	if len(watchFiles) == 0 {
		fatal("no .k files found")
	}

	// Collect directories to watch
	watchDirs := map[string]bool{}
	for _, f := range watchFiles {
		dir := filepath.Dir(f)
		abs, err := filepath.Abs(dir)
		if err != nil {
			abs = dir
		}
		watchDirs[abs] = true
	}
	rtDir := findRuntime()
	if abs, err := filepath.Abs(rtDir); err == nil {
		watchDirs[abs] = true
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fatal("cannot create file watcher: %v", err)
	}
	defer watcher.Close()

	for dir := range watchDirs {
		if err := watcher.Add(dir); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errs.FormatSimple(errs.Warning, fmt.Sprintf("cannot watch %s: %v", dir, err)))
		}
	}

	// Handle Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	var proc *exec.Cmd
	var procMu sync.Mutex
	procDone := make(chan struct{}, 1)

	killProc := func() {
		procMu.Lock()
		defer procMu.Unlock()
		if proc != nil && proc.Process != nil {
			_ = proc.Process.Kill()
			_ = proc.Wait()
			proc = nil
		}
	}

	startProc := func(binPath string) {
		procMu.Lock()
		cmd := exec.Command(binPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		proc = cmd
		procMu.Unlock()

		fmt.Printf("\n%s\n\n", c(errs.BoldCyan, "--- running ---"))

		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errs.FormatSimple(errs.Error, fmt.Sprintf("could not start process: %v", err)))
			return
		}

		go func() {
			_ = cmd.Wait()
			procMu.Lock()
			proc = nil
			procMu.Unlock()
			select {
			case procDone <- struct{}{}:
			default:
			}
		}()
	}

	isRelevant := func(path string) bool {
		ext := filepath.Ext(path)
		return ext == ".k" || ext == ".h" || ext == ".c"
	}

	fmt.Printf("%s %s\n", c(errs.BoldCyan, "kl dev"), c(errs.Dim, "watching for changes..."))
	fmt.Printf("  files: %s\n\n", strings.Join(watchFiles, ", "))

	// Detect hot reload capability (quick parse, no codegen)
	hotReloadMode := checkHotReload(watchFiles)

	if hotReloadMode {
		fmt.Printf("%s\n\n", c(errs.BoldCyan, "hot reload enabled"))
		devHotReload(watchFiles, watcher, sigCh, procDone, killProc, startProc, isRelevant)
	} else {
		devRestart(watchFiles, watcher, sigCh, procDone, killProc, startProc, isRelevant)
	}
}

// devHotReload runs the DLL hot-swap dev loop.
func devHotReload(watchFiles []string, watcher *fsnotify.Watcher, sigCh chan os.Signal,
	procDone chan struct{}, killProc func(), startProc func(string), isRelevant func(string) bool) {

	// Clean up old hot-reload DLL copies
	entries, _ := os.ReadDir("build")
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "_game_") {
			os.Remove(filepath.Join("build", e.Name()))
		}
	}

	// Detect raylib usage
	needsRaylib := false
	for _, f := range watchFiles {
		src, _ := os.ReadFile(f)
		if src != nil && (strings.Contains(string(src), "with rl") || strings.Contains(string(src), "init_window")) {
			needsRaylib = true
		}
	}

	// Build host first (creates import lib that the DLL links against)
	hostRes := buildHost(needsRaylib)
	if !hostRes.ok {
		fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, "failed to build host, falling back to restart mode"))
		devRestart(watchFiles, watcher, sigCh, procDone, killProc, startProc, isRelevant)
		return
	}

	// Build game as DLL (after host, so host.lib import library exists)
	dllRes := doBuild(append([]string{}, watchFiles...), "debug", true)

	// Start host
	if dllRes.ok {
		startProc(hostRes.outputPath)
	} else {
		fmt.Printf("\n%s\n", c(errs.Yellow, "fix errors and save to retry"))
	}

	// Watch loop
	var debounceTimer *time.Timer
	debounceDelay := 200 * time.Millisecond

	rebuildDLL := func() {
		// Check if source still supports hot reload
		if !checkHotReload(watchFiles) {
			// No more hot-reloadable methods — fall back to full restart
			fmt.Printf("\n%s\n", c(errs.Dim, "methods changed, restarting..."))
			killProc()

			res := doBuild(append([]string{}, watchFiles...), "debug", false)
			if res.ok {
				startProc(res.outputPath)
			} else {
				fmt.Printf("\n%s\n", c(errs.Yellow, "fix errors and save to retry"))
			}
			return
		}

		fmt.Printf("\n%s\n\n", c(errs.BoldCyan, "--- hot reloading ---"))

		res := doBuild(append([]string{}, watchFiles...), "debug", true)
		if res.ok {
			reloadSignal := filepath.Join("build", ".reload")
			os.WriteFile(reloadSignal, []byte("reload"), 0644)
			fmt.Printf("%s\n", c(errs.Dim, "signaled host to reload DLL"))
		} else {
			fmt.Printf("\n%s\n", c(errs.Yellow, "fix errors and save to retry"))
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if !isRelevant(event.Name) {
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceDelay, rebuildDLL)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "%s\n", errs.FormatSimple(errs.Warning, fmt.Sprintf("watcher: %v", err)))

		case <-procDone:
			fmt.Printf("\n%s\n", c(errs.Dim, "host exited, waiting for changes..."))

		case <-sigCh:
			fmt.Printf("\n%s\n", c(errs.Dim, "shutting down..."))
			killProc()
			return
		}
	}
}

// devRestart runs the full-restart dev loop (for programs without update/render).
func devRestart(watchFiles []string, watcher *fsnotify.Watcher, sigCh chan os.Signal,
	procDone chan struct{}, killProc func(), startProc func(string), isRelevant func(string) bool) {

	// Initial build & run
	res := doBuild(append([]string{}, watchFiles...), "debug", false)
	if res.ok {
		startProc(res.outputPath)
	} else {
		fmt.Printf("\n%s\n", c(errs.Yellow, "fix errors and save to retry"))
	}

	var debounceTimer *time.Timer
	debounceDelay := 200 * time.Millisecond

	rebuild := func() {
		killProc()
		fmt.Printf("\n%s\n\n", c(errs.BoldCyan, "--- rebuilding ---"))

		res := doBuild(append([]string{}, watchFiles...), "debug", false)
		if res.ok {
			startProc(res.outputPath)
		} else {
			fmt.Printf("\n%s\n", c(errs.Yellow, "fix errors and save to retry"))
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if !isRelevant(event.Name) {
				continue
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceDelay, rebuild)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "%s\n", errs.FormatSimple(errs.Warning, fmt.Sprintf("watcher: %v", err)))

		case <-procDone:
			fmt.Printf("\n%s\n", c(errs.Dim, "process exited, waiting for changes..."))

		case <-sigCh:
			fmt.Printf("\n%s\n", c(errs.Dim, "shutting down..."))
			killProc()
			return
		}
	}
}

// ---------- compile pipeline ----------

// compileFile compiles a .k file to .c. Returns the C file path, whether it uses
// raylib, whether it has hot reload methods, and any error.
func compileFile(path string, dllMode bool) (string, bool, bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, errs.FormatFileError(path, fmt.Sprintf("could not read file: %v", err)))
		return "", false, false, err
	}

	// Lex
	lex := lexer.New(src)
	tokens := lex.Tokenize()

	lexErrors := lex.Errors()
	if len(lexErrors) > 0 {
		for _, le := range lexErrors {
			d := errs.Diagnostic{
				File:    path,
				Line:    le.Line,
				Col:     le.Col,
				EndCol:  le.Col + 1,
				Kind:    errs.Error,
				Message: le.Message,
				Source:  errs.GetSourceLine(src, le.Line),
			}
			fmt.Fprintln(os.Stderr, d.Format())
		}
		return "", false, false, fmt.Errorf("lexer errors in %s", path)
	}

	// Parse
	p := parser.New(tokens)
	p.SetSource(src, path)
	file, err := p.Parse()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return "", false, false, err
	}

	// Semantic checks
	doc := &analysis.Document{URI: path, Source: src, Tokens: tokens, AST: file}
	doc.Gen = codegen.New(file)
	doc.Check()
	if len(doc.Diags) > 0 {
		for _, d := range doc.Diags {
			fmt.Fprintln(os.Stderr, d.Format())
		}
		return "", false, false, fmt.Errorf("semantic errors in %s", path)
	}

	// Generate C
	gen := doc.Gen
	gen.DLLMode = dllMode
	cSource := gen.Generate()
	usesRaylib := gen.UsesRaylib()
	hasHotReload := gen.HasHotReloadMethods()

	base := strings.TrimSuffix(filepath.Base(path), ".k")
	cPath := filepath.Join("build", base+".c")
	if err := os.WriteFile(cPath, []byte(cSource), 0644); err != nil {
		fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, fmt.Sprintf("could not write %s: %v", cPath, err)))
		return "", false, false, err
	}

	fmt.Printf("  %s -> %s\n", path, cPath)
	return cPath, usesRaylib, hasHotReload, nil
}

// ---------- utilities ----------

func isWindows() bool {
	return os.PathSeparator == '\\'
}

func c(code, text string) string {
	return code + text + errs.Reset
}

func findRuntime() string {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	candidates := []string{
		filepath.Join(exeDir, "runtime"),
		filepath.Join(exeDir, "..", "runtime"),
		"runtime",
	}

	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, "kl_runtime.h")); err == nil {
			return dir
		}
	}
	return "runtime"
}

func findRaylib() string {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	candidates := []string{
		filepath.Join(exeDir, "raylib"),
		filepath.Join(exeDir, "..", "raylib"),
		"raylib",
		"C:\\raylib\\raylib",
		"/usr/local",
	}

	if envPath := os.Getenv("RAYLIB_PATH"); envPath != "" {
		candidates = append([]string{envPath}, candidates...)
	}

	for _, dir := range candidates {
		headerPath := filepath.Join(dir, "include", "raylib.h")
		if _, err := os.Stat(headerPath); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "raylib.h")); err == nil {
			return dir
		}
	}
	return ""
}

func findCC() string {
	for _, name := range []string{"cc", "gcc", "clang", "cl"} {
		if _, err := exec.LookPath(name); err == nil {
			return name
		}
	}
	return "gcc"
}

func fatal(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, errs.FormatSimple(errs.Error, msg))
	os.Exit(1)
}
