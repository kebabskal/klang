package codegen

import (
	"fmt"
	"strings"

	"github.com/klang-lang/klang/internal/parser"
)

type scopeVar struct {
	name   string
	cType  string
	isWeak bool
}

// --- Standard library module mappings ---

var stdlibModuleFuncs = map[string]map[string]string{
	"math": {
		"sin":     "sinf",
		"cos":     "cosf",
		"tan":     "tanf",
		"asin":    "asinf",
		"acos":    "acosf",
		"atan":    "atanf",
		"atan2":   "atan2f",
		"sqrt":    "sqrtf",
		"pow":     "powf",
		"abs":     "kl_math_abs",
		"floor":   "floorf",
		"ceil":    "ceilf",
		"round":   "roundf",
		"min":     "kl_math_min",
		"max":     "kl_math_max",
		"clamp":   "kl_math_clamp",
		"lerp":    "kl_math_lerp",
		"sign":    "kl_math_sign",
		"deg2rad": "kl_math_deg2rad",
		"rad2deg": "kl_math_rad2deg",
	},
	"io": {
		"read_file":   "kl_io_read_file",
		"write_file":  "kl_io_write_file",
		"append_file": "kl_io_append_file",
		"file_exists": "kl_io_file_exists",
		"delete_file": "kl_io_delete_file",
		"create_dir":  "kl_io_create_dir",
		"dir_exists":  "kl_io_dir_exists",
		"list_dir":    "kl_io_list_dir",
	},
	"rl": {
		// Window management
		"init_window":          "InitWindow",
		"close_window":         "CloseWindow",
		"window_should_close":  "WindowShouldClose",
		"set_target_fps":       "SetTargetFPS",
		"get_screen_width":     "GetScreenWidth",
		"get_screen_height":    "GetScreenHeight",
		"get_screen_size":      "kl_get_screen_size",
		"toggle_fullscreen":    "ToggleFullscreen",
		"is_window_resized":    "IsWindowResized",
		"set_window_title":     "SetWindowTitle",
		"set_window_size":      "SetWindowSize",
		"get_frame_time":       "GetFrameTime",
		"get_time":             "GetTime",
		"get_fps":              "GetFPS",
		// Drawing
		"begin_drawing":        "BeginDrawing",
		"end_drawing":          "EndDrawing",
		"clear_background":     "ClearBackground",
		"begin_mode_2d":        "BeginMode2D",
		"end_mode_2d":          "EndMode2D",
		"begin_mode_3d":        "BeginMode3D",
		"end_mode_3d":          "EndMode3D",
		// Shapes
		"draw_line":            "DrawLine",
		"draw_line_v":          "kl_draw_line_v",
		"draw_circle":          "DrawCircle",
		"draw_circle_v":        "kl_draw_circle_v",
		"draw_rectangle":       "DrawRectangle",
		"draw_rectangle_v":     "kl_draw_rectangle_v",
		"draw_rectangle_rec":   "DrawRectangleRec",
		"draw_rectangle_lines": "DrawRectangleLines",
		"draw_triangle":        "DrawTriangle",
		// Text
		"draw_text":            "DrawText",
		"draw_text_ex":         "DrawTextEx",
		"measure_text":         "MeasureText",
		"load_font":            "LoadFont",
		"unload_font":          "UnloadFont",
		// Textures
		"load_texture":         "LoadTexture",
		"unload_texture":       "UnloadTexture",
		"draw_texture":         "DrawTexture",
		"draw_texture_v":       "kl_draw_texture_v",
		"draw_texture_ex":      "DrawTextureEx",
		"draw_texture_rec":     "DrawTextureRec",
		"draw_texture_pro":     "kl_draw_texture_pro",
		// Input - keyboard
		"is_key_pressed":       "IsKeyPressed",
		"is_key_down":          "IsKeyDown",
		"is_key_released":      "IsKeyReleased",
		"is_key_up":            "IsKeyUp",
		// Input - mouse
		"is_mouse_button_pressed": "IsMouseButtonPressed",
		"is_mouse_button_down":    "IsMouseButtonDown",
		"is_mouse_button_released":"IsMouseButtonReleased",
		"get_mouse_position":      "kl_get_mouse_position",
		"get_mouse_x":             "GetMouseX",
		"get_mouse_y":             "GetMouseY",
		"get_mouse_wheel_move":    "GetMouseWheelMove",
		// Input - gamepad
		"is_gamepad_available":       "IsGamepadAvailable",
		"is_gamepad_button_pressed":  "IsGamepadButtonPressed",
		"is_gamepad_button_down":     "IsGamepadButtonDown",
		"get_gamepad_axis_movement":  "GetGamepadAxisMovement",
		// Audio
		"init_audio_device":    "InitAudioDevice",
		"close_audio_device":   "CloseAudioDevice",
		"load_sound":           "LoadSound",
		"unload_sound":         "UnloadSound",
		"play_sound":           "PlaySound",
		"stop_sound":           "StopSound",
		"load_music_stream":    "LoadMusicStream",
		"unload_music_stream":  "UnloadMusicStream",
		"play_music_stream":    "PlayMusicStream",
		"stop_music_stream":    "StopMusicStream",
		"update_music_stream":  "UpdateMusicStream",
		"set_master_volume":    "SetMasterVolume",
		// Misc
		"draw_fps":             "DrawFPS",
		"set_exit_key":         "SetExitKey",
		// Helpers
		"color":               "kl_color",
		"color_rgb":           "kl_color_rgb",
		"rect":                "kl_rect",
		"camera2d":            "kl_camera2d",
		"camera3d":            "kl_camera3d",
	},
}

var stdlibModuleConstants = map[string]map[string]string{
	"math": {
		"PI":      "KL_PI",
		"TAU":     "KL_TAU",
		"E":       "KL_E",
		"DEG2RAD": "KL_DEG2RAD",
		"RAD2DEG": "KL_RAD2DEG",
		"INF":     "KL_INF",
		"EPSILON": "KL_EPSILON",
	},
}

// stdlibConstNamespaces maps namespace.Constant → C constant (e.g., Colors.Red → RED)
var stdlibConstNamespaces = map[string]map[string]string{
	"Colors": {
		"LightGray":  "LIGHTGRAY",
		"Gray":       "GRAY",
		"DarkGray":   "DARKGRAY",
		"Yellow":     "YELLOW",
		"Gold":       "GOLD",
		"Orange":     "ORANGE",
		"Pink":       "PINK",
		"Red":        "RED",
		"Maroon":     "MAROON",
		"Green":      "GREEN",
		"Lime":       "LIME",
		"DarkGreen":  "DARKGREEN",
		"SkyBlue":    "SKYBLUE",
		"Blue":       "BLUE",
		"DarkBlue":   "DARKBLUE",
		"Purple":     "PURPLE",
		"Violet":     "VIOLET",
		"DarkPurple": "DARKPURPLE",
		"Beige":      "BEIGE",
		"Brown":      "BROWN",
		"DarkBrown":  "DARKBROWN",
		"White":      "WHITE",
		"Black":      "BLACK",
		"Blank":      "BLANK",
		"Magenta":    "MAGENTA",
		"RayWhite":   "RAYWHITE",
	},
	"Key": {
		"Space":      "KEY_SPACE",
		"Escape":     "KEY_ESCAPE",
		"Enter":      "KEY_ENTER",
		"Tab":        "KEY_TAB",
		"Backspace":  "KEY_BACKSPACE",
		"Delete":     "KEY_DELETE",
		"Right":      "KEY_RIGHT",
		"Left":       "KEY_LEFT",
		"Down":       "KEY_DOWN",
		"Up":         "KEY_UP",
		"LeftShift":  "KEY_LEFT_SHIFT",
		"LeftCtrl":   "KEY_LEFT_CONTROL",
		"LeftAlt":    "KEY_LEFT_ALT",
		"RightShift": "KEY_RIGHT_SHIFT",
		"RightCtrl":  "KEY_RIGHT_CONTROL",
		"RightAlt":   "KEY_RIGHT_ALT",
		"A": "KEY_A", "B": "KEY_B", "C": "KEY_C", "D": "KEY_D",
		"E": "KEY_E", "F": "KEY_F", "G": "KEY_G", "H": "KEY_H",
		"I": "KEY_I", "J": "KEY_J", "K": "KEY_K", "L": "KEY_L",
		"M": "KEY_M", "N": "KEY_N", "O": "KEY_O", "P": "KEY_P",
		"Q": "KEY_Q", "R": "KEY_R", "S": "KEY_S", "T": "KEY_T",
		"U": "KEY_U", "V": "KEY_V", "W": "KEY_W", "X": "KEY_X",
		"Y": "KEY_Y", "Z": "KEY_Z",
		"F1": "KEY_F1", "F2": "KEY_F2", "F3": "KEY_F3", "F4": "KEY_F4",
		"F5": "KEY_F5", "F6": "KEY_F6", "F7": "KEY_F7", "F8": "KEY_F8",
		"F9": "KEY_F9", "F10": "KEY_F10", "F11": "KEY_F11", "F12": "KEY_F12",
		"Zero": "KEY_ZERO", "One": "KEY_ONE", "Two": "KEY_TWO",
		"Three": "KEY_THREE", "Four": "KEY_FOUR", "Five": "KEY_FIVE",
		"Six": "KEY_SIX", "Seven": "KEY_SEVEN", "Eight": "KEY_EIGHT", "Nine": "KEY_NINE",
	},
	"Mouse": {
		"Left":    "MOUSE_BUTTON_LEFT",
		"Right":   "MOUSE_BUTTON_RIGHT",
		"Middle":  "MOUSE_BUTTON_MIDDLE",
	},
	"Gamepad": {
		"LeftStickX":  "GAMEPAD_AXIS_LEFT_X",
		"LeftStickY":  "GAMEPAD_AXIS_LEFT_Y",
		"RightStickX": "GAMEPAD_AXIS_RIGHT_X",
		"RightStickY": "GAMEPAD_AXIS_RIGHT_Y",
		"LeftTrigger":  "GAMEPAD_AXIS_LEFT_TRIGGER",
		"RightTrigger": "GAMEPAD_AXIS_RIGHT_TRIGGER",
	},
}

var stdlibModules = map[string]bool{
	"math": true,
	"io":   true,
	"rl":   true,
}

type Generator struct {
	out     strings.Builder
	indent  int
	file    *parser.File   // primary file (first file, or single file)
	files   []*parser.File // all files (for multi-file compilation)
	classes map[string]*parser.ClassDecl
	// Current context for resolving identifiers
	currentClassName string
	currentClass     *parser.ClassDecl
	localVars        map[string]string         // variable name -> C type (empty string for unknown)
	localVarTypes    map[string]parser.TypeExpr // variable name -> original type expression
	// Generics support
	typeSubstitutions      map[string]string                    // active type param -> C type during monomorphized emission
	genericMethods         map[string]*genericMethodInfo        // "ClassName_methodName" -> info
	genericInstantiations  map[string][]map[string]string       // "ClassName_methodName" -> list of type maps
	genericClasses         map[string]*parser.ClassDecl         // "ClassName" -> generic class template
	genericClassInstances  map[string][]map[string]string       // "ClassName" -> list of type maps
	// Lambda/closure support
	lambdaCounter int
	lambdaDefs    strings.Builder // buffer for lambda capture structs + functions
	// Memory management
	weakFields      map[string]map[string]bool // className -> set of weak field names
	scopeVarStack   [][]scopeVar               // stack of scopes for cleanup tracking
	structLitCounter int
	// Module system (with keyword)
	withModules []string // stack of active "with" modules
	// Raylib detection
	usesRaylib bool
	// DLL hot-reload mode
	DLLMode          bool
	dllInsideMain    bool // when true, we're emitting main() in DLL mode
}

type genericMethodInfo struct {
	className string
	method    *parser.MethodDecl
}

func New(file *parser.File) *Generator {
	g := &Generator{
		file:           file,
		files:          []*parser.File{file},
		classes:        make(map[string]*parser.ClassDecl),
		genericClasses: make(map[string]*parser.ClassDecl),
	}
	for _, cls := range file.Classes {
		g.registerClasses("", cls)
	}
	return g
}

// NewMulti creates a Generator from multiple parsed files.
// All classes from all files are registered, enabling cross-file type resolution.
func NewMulti(files []*parser.File) *Generator {
	if len(files) == 0 {
		return New(&parser.File{})
	}
	g := &Generator{
		file:           files[0],
		files:          files,
		classes:        make(map[string]*parser.ClassDecl),
		genericClasses: make(map[string]*parser.ClassDecl),
	}
	for _, f := range files {
		for _, cls := range f.Classes {
			g.registerClasses("", cls)
		}
	}
	return g
}

// AddFile adds another parsed file's classes to this generator.
// Used to give a single-file generator cross-file type visibility.
func (g *Generator) AddFile(file *parser.File) {
	g.files = append(g.files, file)
	for _, cls := range file.Classes {
		g.registerClasses("", cls)
	}
}

func (g *Generator) registerClasses(prefix string, cls *parser.ClassDecl) {
	fullName := cls.Name
	if prefix != "" {
		fullName = prefix + "_" + cls.Name
	}
	g.classes[fullName] = cls
	for _, nested := range cls.Classes {
		g.registerClasses(fullName, nested)
	}
}

// allClasses returns the top-level classes from all files.
func (g *Generator) allClasses() []*parser.ClassDecl {
	var all []*parser.ClassDecl
	for _, f := range g.files {
		all = append(all, f.Classes...)
	}
	return all
}

// --- Memory management: ownership graph analysis ---

type refEdge struct {
	fieldName   string
	targetClass string
}

func (g *Generator) analyzeOwnershipGraph() {
	g.weakFields = map[string]map[string]bool{}

	// Build adjacency list: className -> list of edges to other classes
	graph := map[string][]refEdge{}
	for name, cls := range g.classes {
		for _, field := range cls.Fields {
			cType := g.fieldCType(field, name)
			targetClass := g.rcTargetClass(cType)
			if targetClass != "" {
				graph[name] = append(graph[name], refEdge{field.Name, targetClass})
			}
			// Also check List<ClassName>
			if field.TypeExpr != nil {
				if gt, ok := field.TypeExpr.(*parser.GenericType); ok && gt.Name == "List" && len(gt.TypeArgs) > 0 {
					elemCType := g.typeToC(gt.TypeArgs[0], name)
					elemTarget := g.rcTargetClass(elemCType)
					if elemTarget != "" {
						graph[name] = append(graph[name], refEdge{field.Name, elemTarget})
					}
				}
			}
		}
	}

	// Find SCCs using Tarjan's algorithm
	sccs := g.tarjanSCC(graph)

	// For each SCC, mark cycle-forming fields as weak
	for _, scc := range sccs {
		sccSet := map[string]bool{}
		for _, name := range scc {
			sccSet[name] = true
		}
		if len(scc) == 1 {
			// Self-referencing check
			name := scc[0]
			for _, e := range graph[name] {
				if e.targetClass == name {
					g.markWeakField(name, e.fieldName)
				}
			}
		} else {
			// Multi-node SCC: mark inter-member fields as weak
			for _, name := range scc {
				for _, e := range graph[name] {
					if sccSet[e.targetClass] {
						g.markWeakField(name, e.fieldName)
					}
				}
			}
		}
	}
}

func (g *Generator) markWeakField(className, fieldName string) {
	if g.weakFields[className] == nil {
		g.weakFields[className] = map[string]bool{}
	}
	g.weakFields[className][fieldName] = true
}

func (g *Generator) tarjanSCC(graph map[string][]refEdge) [][]string {
	index := 0
	stack := []string{}
	onStack := map[string]bool{}
	indices := map[string]int{}
	lowlinks := map[string]int{}
	var sccs [][]string

	var strongconnect func(v string)
	strongconnect = func(v string) {
		indices[v] = index
		lowlinks[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		for _, e := range graph[v] {
			w := e.targetClass
			if _, visited := indices[w]; !visited {
				strongconnect(w)
				if lowlinks[w] < lowlinks[v] {
					lowlinks[v] = lowlinks[w]
				}
			} else if onStack[w] {
				if indices[w] < lowlinks[v] {
					lowlinks[v] = indices[w]
				}
			}
		}

		if lowlinks[v] == indices[v] {
			var scc []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				scc = append(scc, w)
				if w == v {
					break
				}
			}
			sccs = append(sccs, scc)
		}
	}

	for v := range graph {
		if _, visited := indices[v]; !visited {
			strongconnect(v)
		}
	}
	return sccs
}

// fieldCType returns the C type for a field, handling inferred types.
func (g *Generator) fieldCType(field *parser.FieldDecl, className string) string {
	if field.Inferred {
		return g.inferCType(field.Value)
	}
	if field.TypeExpr != nil {
		return g.typeToC(field.TypeExpr, className)
	}
	return "int"
}

// isRefCountedType returns true if the C type represents a heap-allocated refcounted object.
func (g *Generator) isRefCountedType(cType string) bool {
	switch cType {
	case "int", "float", "bool", "const char*", "vec2", "void", "void*", "":
		return false
	}
	return strings.HasSuffix(cType, "*")
}

// rcTargetClass extracts the class name from a refcounted pointer type, or "" if not RC.
func (g *Generator) rcTargetClass(cType string) string {
	if !g.isRefCountedType(cType) {
		return ""
	}
	return strings.TrimSuffix(cType, "*")
}

// isWeakField returns true if the given field on the given class is auto-weak.
func (g *Generator) isWeakField(className, fieldName string) bool {
	if g.weakFields == nil {
		return false
	}
	if m, ok := g.weakFields[className]; ok {
		return m[fieldName]
	}
	return false
}

// --- Scope tracking for cleanup ---

func (g *Generator) pushScope() {
	g.scopeVarStack = append(g.scopeVarStack, []scopeVar{})
}

func (g *Generator) popScope() []scopeVar {
	if len(g.scopeVarStack) == 0 {
		return nil
	}
	n := len(g.scopeVarStack)
	vars := g.scopeVarStack[n-1]
	g.scopeVarStack = g.scopeVarStack[:n-1]
	return vars
}

func (g *Generator) pushScopeVar(name, cType string) {
	if len(g.scopeVarStack) == 0 {
		return
	}
	n := len(g.scopeVarStack)
	g.scopeVarStack[n-1] = append(g.scopeVarStack[n-1], scopeVar{name: name, cType: cType})
}

func (g *Generator) emitScopeCleanup(excludeVar string) {
	for i := len(g.scopeVarStack) - 1; i >= 0; i-- {
		for _, v := range g.scopeVarStack[i] {
			if v.name == excludeVar {
				continue
			}
			g.writeln("kl_release(%s);", v.name)
		}
	}
}

// emitScopeCleanupCurrentOnly releases only the vars in the current (innermost) scope.
func (g *Generator) emitScopeCleanupCurrentOnly() {
	if len(g.scopeVarStack) == 0 {
		return
	}
	vars := g.scopeVarStack[len(g.scopeVarStack)-1]
	for _, v := range vars {
		g.writeln("kl_release(%s);", v.name)
	}
}

// --- Destructor and tracer emission ---

func (g *Generator) emitDestructors(prefix string, cls *parser.ClassDecl) {
	if len(cls.TypeParams) > 0 {
		return
	}
	name := g.cName(prefix, cls.Name)
	g.emitDestructor(name, cls)
	for _, nested := range cls.Classes {
		g.emitDestructors(name, nested)
	}
}

func (g *Generator) emitDestructor(name string, cls *parser.ClassDecl) {
	g.writeln("static void %s_destroy(KlObject* _obj) {", name)
	g.indent++
	g.writeln("%s* self = (%s*)_obj;", name, name)

	prefix := g.extractPrefix(name, cls.Name)
	for _, field := range cls.Fields {
		if cls.Parent != "" && g.isFieldInParent(prefix, cls.Parent, field.Name) {
			continue
		}
		cType := g.fieldCType(field, name)
		if g.isWeakField(name, field.Name) {
			g.writeln("kl_weak_release(&self->%s);", field.Name)
		} else if g.isRefCountedType(cType) {
			g.writeln("kl_release(self->%s);", field.Name)
		}
	}

	if cls.Parent != "" {
		parentName := g.resolveParentName(prefix, cls.Parent)
		g.writeln("%s_destroy((KlObject*)&self->_base);", parentName)
	}

	g.indent--
	g.writeln("}")
	g.writeln("")
}

func (g *Generator) emitTracers(prefix string, cls *parser.ClassDecl) {
	if len(cls.TypeParams) > 0 {
		return
	}
	name := g.cName(prefix, cls.Name)
	g.emitTracer(name, cls)
	for _, nested := range cls.Classes {
		g.emitTracers(name, nested)
	}
}

func (g *Generator) emitTracer(name string, cls *parser.ClassDecl) {
	g.writeln("static void %s_trace(KlObject* _obj, void (*visit)(KlObject*)) {", name)
	g.indent++
	g.writeln("%s* self = (%s*)_obj;", name, name)

	prefix := g.extractPrefix(name, cls.Name)
	for _, field := range cls.Fields {
		if cls.Parent != "" && g.isFieldInParent(prefix, cls.Parent, field.Name) {
			continue
		}
		cType := g.fieldCType(field, name)
		if !g.isWeakField(name, field.Name) && g.isRefCountedType(cType) {
			g.writeln("if (self->%s) visit((KlObject*)self->%s);", field.Name, field.Name)
		}
	}

	if cls.Parent != "" {
		parentName := g.resolveParentName(prefix, cls.Parent)
		g.writeln("%s_trace((KlObject*)&self->_base, visit);", parentName)
	}

	g.indent--
	g.writeln("}")
	g.writeln("")
}

// --- Destructor/tracer prototypes ---

func (g *Generator) emitDestructorPrototypes(prefix string, cls *parser.ClassDecl) {
	if len(cls.TypeParams) > 0 {
		return
	}
	name := g.cName(prefix, cls.Name)
	g.writeln("static void %s_destroy(KlObject* _obj);", name)
	g.writeln("static void %s_trace(KlObject* _obj, void (*visit)(KlObject*));", name)
	for _, nested := range cls.Classes {
		g.emitDestructorPrototypes(name, nested)
	}
}

// scanForRaylib walks the AST to detect if the program uses raylib
// (rl module, Colors/Key/Mouse constants, or with rl).
func (g *Generator) scanForRaylib() {
	for _, cls := range g.allClasses() {
		for _, m := range cls.Methods {
			if m.Body != nil {
				if g.scanStmtsForRaylib(m.Body.Stmts) {
					return
				}
			}
		}
		if cls.Constructor != nil && cls.Constructor.Body != nil {
			if g.scanStmtsForRaylib(cls.Constructor.Body.Stmts) {
				return
			}
		}
	}
}

func (g *Generator) scanStmtsForRaylib(stmts []parser.Stmt) bool {
	for _, stmt := range stmts {
		if g.scanStmtForRaylib(stmt) {
			g.usesRaylib = true
			return true
		}
	}
	return false
}

func (g *Generator) scanStmtForRaylib(stmt parser.Stmt) bool {
	switch s := stmt.(type) {
	case *parser.WithStmt:
		if s.Module == "rl" {
			return true
		}
		if s.Body != nil {
			return g.scanStmtsForRaylib(s.Body.Stmts)
		}
	case *parser.ExprStmt:
		return g.scanExprForRaylib(s.Expr)
	case *parser.VarDecl:
		return g.scanExprForRaylib(s.Value)
	case *parser.AssignStmt:
		return g.scanExprForRaylib(s.Target) || g.scanExprForRaylib(s.Value)
	case *parser.IfStmt:
		if g.scanExprForRaylib(s.Condition) {
			return true
		}
		if s.Then != nil && g.scanStmtsForRaylib(s.Then.Stmts) {
			return true
		}
		if s.Else != nil {
			return g.scanStmtForRaylib(s.Else)
		}
		if s.ThenStmt != nil {
			return g.scanStmtForRaylib(s.ThenStmt)
		}
	case *parser.ForStmt:
		if g.scanExprForRaylib(s.Iterable) {
			return true
		}
		if s.Body != nil {
			return g.scanStmtsForRaylib(s.Body.Stmts)
		}
	case *parser.WhileStmt:
		if g.scanExprForRaylib(s.Condition) {
			return true
		}
		if s.Body != nil {
			return g.scanStmtsForRaylib(s.Body.Stmts)
		}
	case *parser.ReturnStmt:
		if s.Value != nil {
			return g.scanExprForRaylib(s.Value)
		}
	case *parser.Block:
		return g.scanStmtsForRaylib(s.Stmts)
	}
	return false
}

func (g *Generator) scanExprForRaylib(expr parser.Expr) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *parser.MemberExpr:
		// Check for rl.xxx, Colors.xxx, Key.xxx, Mouse.xxx
		if ident, ok := e.Object.(*parser.Ident); ok {
			switch ident.Name {
			case "rl", "Colors", "Key", "Mouse", "Gamepad":
				return true
			}
		}
		return g.scanExprForRaylib(e.Object)
	case *parser.CallExpr:
		if g.scanExprForRaylib(e.Callee) {
			return true
		}
		for _, arg := range e.Args {
			if g.scanExprForRaylib(arg) {
				return true
			}
		}
	case *parser.BinaryExpr:
		return g.scanExprForRaylib(e.Left) || g.scanExprForRaylib(e.Right)
	case *parser.UnaryExpr:
		return g.scanExprForRaylib(e.Operand)
	case *parser.IndexExpr:
		return g.scanExprForRaylib(e.Object) || g.scanExprForRaylib(e.Index)
	}
	return false
}

// UsesRaylib returns true if the program uses the rl module (call after Generate).
func (g *Generator) UsesRaylib() bool {
	return g.usesRaylib
}

func (g *Generator) Generate() string {
	// Scan AST to detect raylib usage
	g.scanForRaylib()

	if g.usesRaylib {
		g.writeln("#define KL_USE_RAYLIB")
	}
	g.writeln("#include \"kl_runtime.h\"")
	g.writeln("")

	// Analyze ownership graph for auto-weak detection
	g.analyzeOwnershipGraph()

	for _, cls := range g.allClasses() {
		g.emitForwardDecls("", cls)
	}
	g.writeln("")

	for _, cls := range g.allClasses() {
		g.emitStructDefs("", cls)
	}

	g.emitTypeIds()

	// Emit destructor/tracer forward declarations (needed before constructors)
	for _, cls := range g.allClasses() {
		g.emitDestructorPrototypes("", cls)
	}
	g.writeln("")

	// Emit destructor and tracer implementations
	for _, cls := range g.allClasses() {
		g.emitDestructors("", cls)
	}
	for _, cls := range g.allClasses() {
		g.emitTracers("", cls)
	}

	for _, cls := range g.allClasses() {
		g.emitEnumDefs("", cls)
	}

	// Pre-scan: register generics and collect instantiations (two-pass)
	g.genericMethods = map[string]*genericMethodInfo{}
	g.genericInstantiations = map[string][]map[string]string{}
	g.genericClasses = map[string]*parser.ClassDecl{}
	g.genericClassInstances = map[string][]map[string]string{}
	// Pass 1: register all generic templates (methods and classes)
	for _, cls := range g.allClasses() {
		g.registerGenericTemplates("", cls)
	}
	// Pass 2: scan method bodies for calls to generic methods/classes
	for _, cls := range g.allClasses() {
		g.scanForGenericCalls("", cls)
	}
	// Emit monomorphized generic class structs, type IDs, prototypes, and implementations
	g.emitGenericClassCode()

	for _, cls := range g.allClasses() {
		g.emitPrototypes("", cls)
	}
	g.writeln("")

	// Save position before implementations — lambda defs will be inserted here
	preImplLen := g.out.Len()

	for _, cls := range g.allClasses() {
		g.emitImplementations("", cls)
	}

	g.emitEntryPoint()

	// Insert lambda capture structs and functions before implementations
	if g.lambdaDefs.Len() > 0 {
		result := g.out.String()
		return result[:preImplLen] + g.lambdaDefs.String() + result[preImplLen:]
	}

	return g.out.String()
}

// --- Generics support ---

// Pass 1: register all generic method and class templates
func (g *Generator) registerGenericTemplates(prefix string, cls *parser.ClassDecl) {
	name := g.cName(prefix, cls.Name)

	if len(cls.TypeParams) > 0 {
		g.genericClasses[name] = cls
		return // don't recurse into generic class internals
	}

	for _, method := range cls.Methods {
		if len(method.TypeParams) > 0 {
			key := name + "_" + method.Name
			g.genericMethods[key] = &genericMethodInfo{className: name, method: method}
		}
	}

	for _, nested := range cls.Classes {
		g.registerGenericTemplates(name, nested)
	}
}

// Pass 2: scan method bodies for calls to generic methods/classes
func (g *Generator) scanForGenericCalls(prefix string, cls *parser.ClassDecl) {
	if len(cls.TypeParams) > 0 {
		return
	}
	name := g.cName(prefix, cls.Name)

	for _, method := range cls.Methods {
		if len(method.TypeParams) == 0 && method.Body != nil {
			g.withClass(name, cls, func() {
				g.localVars = map[string]string{}
				g.localVarTypes = map[string]parser.TypeExpr{}
				for _, p := range method.Params {
					g.localVars[p.Name] = g.typeToC(p.TypeExpr, name)
				}
				g.scanBlockForGenericCalls(method.Body, name)
				g.localVars = nil
				g.localVarTypes = nil
			})
		}
	}

	if cls.Constructor != nil && cls.Constructor.Body != nil {
		g.withClass(name, cls, func() {
			g.scanBlockForGenericCalls(cls.Constructor.Body, name)
		})
	}

	for _, nested := range cls.Classes {
		g.scanForGenericCalls(name, nested)
	}
}

func (g *Generator) scanBlockForGenericCalls(block *parser.Block, className string) {
	for _, stmt := range block.Stmts {
		g.scanStmtForGenericCalls(stmt, className)
	}
}

func (g *Generator) scanStmtForGenericCalls(stmt parser.Stmt, className string) {
	switch s := stmt.(type) {
	case *parser.ExprStmt:
		g.scanExprForGenericCalls(s.Expr, className)
	case *parser.VarDecl:
		if s.Value != nil {
			// Register the local var type for inference
			if g.localVars != nil {
				cType := g.inferCType(s.Value)
				if s.TypeExpr != nil {
					cType = g.typeToC(s.TypeExpr, className)
				}
				g.localVars[s.Name] = cType
			}
			if g.localVarTypes != nil && s.TypeExpr != nil {
				g.localVarTypes[s.Name] = s.TypeExpr
			}
			g.scanExprForGenericCalls(s.Value, className)
		}
	case *parser.ReturnStmt:
		if s.Value != nil {
			g.scanExprForGenericCalls(s.Value, className)
		}
	case *parser.IfStmt:
		g.scanExprForGenericCalls(s.Condition, className)
		if s.Then != nil {
			g.scanBlockForGenericCalls(s.Then, className)
		}
		if s.Else != nil {
			g.scanStmtForGenericCalls(s.Else, className)
		}
	case *parser.ForStmt:
		g.scanExprForGenericCalls(s.Iterable, className)
		if s.Body != nil {
			g.scanBlockForGenericCalls(s.Body, className)
		}
	case *parser.WhileStmt:
		g.scanExprForGenericCalls(s.Condition, className)
		if s.Body != nil {
			g.scanBlockForGenericCalls(s.Body, className)
		}
	case *parser.AssignStmt:
		g.scanExprForGenericCalls(s.Value, className)
	case *parser.Block:
		g.scanBlockForGenericCalls(s, className)
	}
}

func (g *Generator) scanExprForGenericCalls(expr parser.Expr, className string) {
	call, ok := expr.(*parser.CallExpr)
	if !ok {
		return
	}
	// Scan args recursively
	for _, arg := range call.Args {
		g.scanExprForGenericCalls(arg, className)
	}

	ident, ok := call.Callee.(*parser.Ident)
	if !ok {
		return
	}

	// Check if this is a call to a generic method
	key := className + "_" + ident.Name
	if info, ok := g.genericMethods[key]; ok {
		typeMap := g.resolveGenericTypeArgs(info, call, className)
		g.registerInstantiation(key, typeMap)
		return
	}

	// Check if this is a generic class constructor call
	// Look up in all registered generic classes (could be nested or top-level)
	gcName := g.resolveGenericClassName(ident.Name, className)
	if gcls, ok := g.genericClasses[gcName]; ok {
		typeMap := g.resolveGenericClassTypeArgs(gcls, call, className)
		g.registerClassInstantiation(gcName, typeMap, gcls.TypeParams)
	}
}

func (g *Generator) resolveGenericTypeArgs(info *genericMethodInfo, call *parser.CallExpr, className string) map[string]string {
	typeMap := map[string]string{}
	if len(call.TypeArgs) > 0 {
		// Explicit: <int, string>
		for i, tp := range info.method.TypeParams {
			if i < len(call.TypeArgs) {
				typeMap[tp] = g.typeToC(call.TypeArgs[i], className)
			}
		}
	} else {
		// Implicit: infer from arguments
		typeParamSet := map[string]bool{}
		for _, tp := range info.method.TypeParams {
			typeParamSet[tp] = true
		}
		for i, param := range info.method.Params {
			if i >= len(call.Args) {
				break
			}
			if st, ok := param.TypeExpr.(*parser.SimpleType); ok && typeParamSet[st.Name] {
				typeMap[st.Name] = g.inferCType(call.Args[i])
			}
		}
	}
	return typeMap
}

func (g *Generator) registerInstantiation(key string, typeMap map[string]string) {
	// Deduplicate
	mangledName := g.mangledSuffix(typeMap, g.genericMethods[key].method.TypeParams)
	for _, existing := range g.genericInstantiations[key] {
		if g.mangledSuffix(existing, g.genericMethods[key].method.TypeParams) == mangledName {
			return
		}
	}
	g.genericInstantiations[key] = append(g.genericInstantiations[key], typeMap)
}

func (g *Generator) resolveGenericClassName(name, context string) string {
	// Direct match
	if _, ok := g.genericClasses[name]; ok {
		return name
	}
	// Try with context prefix (nested generic class)
	if context != "" {
		full := context + "_" + name
		if _, ok := g.genericClasses[full]; ok {
			return full
		}
	}
	// Search all generic classes for suffix match
	for fullName := range g.genericClasses {
		if strings.HasSuffix(fullName, "_"+name) {
			return fullName
		}
	}
	return name
}

func (g *Generator) resolveGenericClassTypeArgs(cls *parser.ClassDecl, call *parser.CallExpr, context string) map[string]string {
	typeMap := map[string]string{}
	if len(call.TypeArgs) > 0 {
		// Explicit: GenericClass<int>(10)
		for i, tp := range cls.TypeParams {
			if i < len(call.TypeArgs) {
				typeMap[tp] = g.typeToC(call.TypeArgs[i], context)
			}
		}
	} else if cls.Constructor != nil {
		// Implicit: GenericClass("value") — infer from constructor params
		typeParamSet := map[string]bool{}
		for _, tp := range cls.TypeParams {
			typeParamSet[tp] = true
		}
		for i, param := range cls.Constructor.Params {
			if i >= len(call.Args) {
				break
			}
			if st, ok := param.TypeExpr.(*parser.SimpleType); ok && typeParamSet[st.Name] {
				typeMap[st.Name] = g.inferCType(call.Args[i])
			}
		}
	}
	return typeMap
}

func (g *Generator) registerClassInstantiation(key string, typeMap map[string]string, typeParams []string) {
	mangledName := g.mangledSuffix(typeMap, typeParams)
	for _, existing := range g.genericClassInstances[key] {
		if g.mangledSuffix(existing, typeParams) == mangledName {
			return
		}
	}
	g.genericClassInstances[key] = append(g.genericClassInstances[key], typeMap)
}

func mangleCType(cType string) string {
	switch cType {
	case "int":
		return "int"
	case "float":
		return "float"
	case "bool":
		return "bool"
	case "const char*":
		return "string"
	case "vec2":
		return "vec2"
	default:
		return strings.TrimSuffix(cType, "*")
	}
}

func (g *Generator) mangledSuffix(typeMap map[string]string, typeParams []string) string {
	var parts []string
	for _, tp := range typeParams {
		parts = append(parts, mangleCType(typeMap[tp]))
	}
	return strings.Join(parts, "_")
}

func (g *Generator) mangledGenericName(className, methodName string, typeMap map[string]string, typeParams []string) string {
	return className + "_" + methodName + "_" + g.mangledSuffix(typeMap, typeParams)
}

// HasHotReloadMethods returns true if the main class has non-main methods
// that can be hot-reloaded via function pointer indirection.
func (g *Generator) HasHotReloadMethods() bool {
	for _, cls := range g.allClasses() {
		hasMain := false
		hasOther := false
		for _, method := range cls.Methods {
			if method.Name == "main" && len(method.Params) == 0 {
				hasMain = true
			} else {
				hasOther = true
			}
		}
		if hasMain && hasOther {
			return true
		}
	}
	return false
}

func (g *Generator) emitEntryPoint() {
	// Find a main() method in any class and generate a C main()
	for _, cls := range g.allClasses() {
		className := cls.Name
		for _, method := range cls.Methods {
			if method.Name == "main" && len(method.Params) == 0 {
				if g.DLLMode {
					g.emitDLLHooks(cls, className)
				} else {
					g.writeln("int main(int argc, char** argv) {")
					g.indent++
					g.writeln("%s* _instance = %s_new(%s);", className, className, g.defaultConstructorArgs(cls))
					g.writeln("%s_main(_instance);", className)
					g.writeln("kl_release(_instance);")
					g.writeln("return 0;")
					g.indent--
					g.writeln("}")
				}
				return
			}
		}
	}
}

func (g *Generator) emitDLLHooks(cls *parser.ClassDecl, className string) {
	// Platform-specific export macro
	g.writeln("#ifdef _WIN32")
	g.writeln("#define KL_EXPORT __declspec(dllexport)")
	g.writeln("#else")
	g.writeln("#define KL_EXPORT __attribute__((visibility(\"default\")))")
	g.writeln("#endif")
	g.writeln("")

	// game_create — allocates instance (constructor inits fn pointers)
	g.writeln("KL_EXPORT void* game_create(void) {")
	g.indent++
	g.writeln("return %s_new(%s);", className, g.defaultConstructorArgs(cls))
	g.indent--
	g.writeln("}")
	g.writeln("")

	// game_main — runs the full main() method (called in a thread by host)
	g.writeln("KL_EXPORT void game_main(void* _self) {")
	g.indent++
	g.writeln("%s_main((%s*)_self);", className, className)
	g.indent--
	g.writeln("}")
	g.writeln("")

	// game_patch — updates function pointers to this DLL's implementations
	g.writeln("KL_EXPORT void game_patch(void* _self) {")
	g.indent++
	g.writeln("%s* self = (%s*)_self;", className, className)
	for _, m := range cls.Methods {
		if m.Name == "main" {
			continue
		}
		g.writeln("self->_fn_%s = %s_%s;", m.Name, className, m.Name)
	}
	g.indent--
	g.writeln("}")
	g.writeln("")

	// game_destroy — releases instance
	g.writeln("KL_EXPORT void game_destroy(void* _self) {")
	g.indent++
	g.writeln("if (_self) kl_release(_self);")
	g.indent--
	g.writeln("}")
	g.writeln("")
}

func (g *Generator) defaultConstructorArgs(cls *parser.ClassDecl) string {
	if cls.Constructor == nil || len(cls.Constructor.Params) == 0 {
		return ""
	}
	// Generate default values for constructor params
	args := make([]string, len(cls.Constructor.Params))
	for i, p := range cls.Constructor.Params {
		switch g.typeToC(p.TypeExpr, cls.Name) {
		case "int":
			args[i] = "0"
		case "float":
			args[i] = "0.0f"
		case "bool":
			args[i] = "false"
		case "const char*":
			args[i] = "\"\""
		default:
			args[i] = "NULL"
		}
	}
	return strings.Join(args, ", ")
}

func (g *Generator) cName(prefix, name string) string {
	if prefix != "" {
		return prefix + "_" + name
	}
	return name
}

// --- Identifier resolution ---

func (g *Generator) isField(name string) bool {
	if g.currentClass == nil {
		return false
	}
	return g.findFieldInClass(g.currentClass, name)
}

func (g *Generator) findFieldInClass(cls *parser.ClassDecl, name string) bool {
	for _, f := range cls.Fields {
		if f.Name == name {
			return true
		}
	}
	// Check parent class fields
	if cls.Parent != "" {
		parentName := g.resolveParentName(g.currentClassName, cls.Parent)
		if parent, ok := g.classes[parentName]; ok {
			return g.findFieldInClass(parent, name)
		}
	}
	return false
}

func (g *Generator) isProperty(name string) bool {
	if g.currentClass == nil {
		return false
	}
	for _, p := range g.currentClass.Properties {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (g *Generator) isListExpr(expr parser.Expr) bool {
	if ident, ok := expr.(*parser.Ident); ok {
		// Check local variable types
		if g.localVars != nil {
			if t, ok := g.localVars[ident.Name]; ok {
				return t == "KlList*"
			}
		}
		// Check local var type expressions
		if g.localVarTypes != nil {
			if typeExpr, ok := g.localVarTypes[ident.Name]; ok {
				if gt, ok := typeExpr.(*parser.GenericType); ok && gt.Name == "List" {
					return true
				}
			}
		}
		// Check class fields
		if g.currentClass != nil {
			for _, f := range g.currentClass.Fields {
				if f.Name == ident.Name {
					if gt, ok := f.TypeExpr.(*parser.GenericType); ok && gt.Name == "List" {
						return true
					}
				}
			}
		}
	}
	return false
}

func (g *Generator) isMethod(name string) bool {
	if g.currentClass == nil {
		return false
	}
	return g.findMethodInClass(g.currentClass, name)
}

func (g *Generator) findMethodInClass(cls *parser.ClassDecl, name string) bool {
	for _, m := range cls.Methods {
		if m.Name == name {
			return true
		}
	}
	if cls.Parent != "" {
		parentName := g.resolveParentName(g.currentClassName, cls.Parent)
		if parent, ok := g.classes[parentName]; ok {
			return g.findMethodInClass(parent, name)
		}
	}
	return false
}

// --- Forward declarations ---

func (g *Generator) emitForwardDecls(prefix string, cls *parser.ClassDecl) {
	if len(cls.TypeParams) > 0 {
		return // generic template — monomorphized versions emitted separately
	}
	name := g.cName(prefix, cls.Name)
	g.writeln("typedef struct %s %s;", name, name)
	for _, nested := range cls.Classes {
		g.emitForwardDecls(name, nested)
	}
}

// --- Struct definitions ---

func (g *Generator) emitStructDefs(prefix string, cls *parser.ClassDecl) {
	if len(cls.TypeParams) > 0 {
		return // generic template — monomorphized versions emitted separately
	}
	name := g.cName(prefix, cls.Name)

	g.writeln("struct %s {", name)
	g.indent++

	if cls.Parent != "" {
		parentName := g.resolveParentName(prefix, cls.Parent)
		g.writeln("%s _base;", parentName)
	} else {
		g.writeln("KlHeader _header;")
	}

	// Only emit fields that aren't overrides from the parent
	for _, field := range cls.Fields {
		if cls.Parent != "" && g.isFieldInParent(prefix, cls.Parent, field.Name) {
			continue // skip, it's inherited via _base
		}
		cType := g.typeToC(field.TypeExpr, name)
		if field.Inferred {
			cType = g.inferCType(field.Value)
		}
		if cType == "void" {
			continue // skip void fields
		}
		// Weak fields are stored as KlWeakSlot*
		if g.isWeakField(name, field.Name) {
			g.writeln("KlWeakSlot* %s;", field.Name)
		} else {
			g.writeln("%s %s;", cType, field.Name)
		}
	}

	// In DLL mode, add function pointer fields for hot-reloadable methods
	if g.DLLMode {
		for _, m := range cls.Methods {
			if m.Name == "main" {
				continue
			}
			retType := g.returnTypeToC(m.ReturnType, name)
			// Build parameter type list: (ClassName*, param types...)
			paramTypes := fmt.Sprintf("struct %s*", name)
			for _, p := range m.Params {
				paramTypes += ", " + g.typeToC(p.TypeExpr, name)
			}
			g.writeln("%s (*_fn_%s)(%s);", retType, m.Name, paramTypes)
		}
	}

	g.indent--
	g.writeln("};")
	g.writeln("")

	for _, nested := range cls.Classes {
		g.emitStructDefs(name, nested)
	}
}

func (g *Generator) isFieldInParent(prefix, parent, fieldName string) bool {
	parentName := g.resolveParentName(prefix, parent)
	if pcls, ok := g.classes[parentName]; ok {
		for _, f := range pcls.Fields {
			if f.Name == fieldName {
				return true
			}
		}
	}
	return false
}

// --- Enum definitions ---

func (g *Generator) emitEnumDefs(prefix string, cls *parser.ClassDecl) {
	if len(cls.TypeParams) > 0 {
		return
	}
	name := g.cName(prefix, cls.Name)
	for _, en := range cls.Enums {
		enumName := g.cName(name, en.Name)
		g.writeln("typedef enum {")
		g.indent++
		for _, member := range en.Members {
			memberName := g.cName(enumName, member.Name)
			if member.Value != nil {
				g.writeln("%s = %s,", memberName, g.exprToC(member.Value))
			} else {
				g.writeln("%s,", memberName)
			}
		}
		g.indent--
		g.writeln("} %s;", enumName)
		g.writeln("")
	}

	for _, nested := range cls.Classes {
		g.emitEnumDefs(name, nested)
	}
}

// --- Prototypes ---

func (g *Generator) emitPrototypes(prefix string, cls *parser.ClassDecl) {
	if len(cls.TypeParams) > 0 {
		return // generic template — monomorphized versions emitted separately
	}
	name := g.cName(prefix, cls.Name)

	g.writeln("%s* %s_new(%s);", name, name, g.constructorParams(cls, name))

	for _, method := range cls.Methods {
		if len(method.TypeParams) > 0 {
			// Emit prototypes for each monomorphized instantiation
			key := name + "_" + method.Name
			for _, typeMap := range g.genericInstantiations[key] {
				g.typeSubstitutions = typeMap
				mangledName := g.mangledGenericName(name, method.Name, typeMap, method.TypeParams)
				retType := g.returnTypeToC(method.ReturnType, name)
				params := g.methodParams(name, method)
				g.writeln("%s %s(%s);", retType, mangledName, params)
				g.typeSubstitutions = nil
			}
			continue
		}
		retType := g.returnTypeToC(method.ReturnType, name)
		params := g.methodParams(name, method)
		g.writeln("%s %s_%s(%s);", retType, name, method.Name, params)
	}

	for _, prop := range cls.Properties {
		propType := g.typeToC(prop.TypeExpr, name)
		if prop.Getter != nil {
			g.writeln("%s %s_get_%s(%s* self);", propType, name, prop.Name, name)
		}
		if prop.Setter != nil {
			g.writeln("void %s_set_%s(%s* self, %s value);", name, prop.Name, name, propType)
		}
	}

	for _, nested := range cls.Classes {
		g.emitPrototypes(name, nested)
	}
}

// --- Implementations ---

func (g *Generator) emitImplementations(prefix string, cls *parser.ClassDecl) {
	if len(cls.TypeParams) > 0 {
		return // generic template — monomorphized versions emitted separately
	}
	name := g.cName(prefix, cls.Name)

	g.emitConstructor(name, cls)

	for _, method := range cls.Methods {
		if len(method.TypeParams) > 0 {
			// Emit monomorphized instantiations
			key := name + "_" + method.Name
			for _, typeMap := range g.genericInstantiations[key] {
				g.withClass(name, cls, func() {
					g.emitGenericInstantiation(name, method, typeMap)
				})
			}
			continue
		}
		g.withClass(name, cls, func() {
			g.emitMethod(name, method)
		})
	}

	for _, prop := range cls.Properties {
		g.withClass(name, cls, func() {
			g.emitProperty(name, prop)
		})
	}

	for _, nested := range cls.Classes {
		g.emitImplementations(name, nested)
	}
}

func (g *Generator) withClass(name string, cls *parser.ClassDecl, fn func()) {
	prevName := g.currentClassName
	prevClass := g.currentClass
	g.currentClassName = name
	g.currentClass = cls
	fn()
	g.currentClassName = prevName
	g.currentClass = prevClass
}

func (g *Generator) emitConstructor(name string, cls *parser.ClassDecl) {
	params := g.constructorParams(cls, name)
	g.writeln("%s* %s_new(%s) {", name, name, params)
	g.indent++
	g.writeln("%s* self = (%s*)kl_alloc_rc(sizeof(%s), %s_destroy);", name, name, name, name)

	// Set type ID and tracer
	if cls.Parent != "" {
		g.writeln("self->_base._header.type_id = KLTYPE_%s;", name)
		g.writeln("self->_base._header.tracer = %s_trace;", name)
	} else {
		g.writeln("self->_header.type_id = KLTYPE_%s;", name)
		g.writeln("self->_header.tracer = %s_trace;", name)
	}

	for _, field := range cls.Fields {
		if field.Value != nil {
			target := "self->" + field.Name
			if cls.Parent != "" && g.isFieldInParent("", cls.Parent, field.Name) {
				target = "self->_base." + field.Name
			}

			// Array literals: create list then push each element
			if arr, ok := field.Value.(*parser.ArrayLit); ok {
				itemsRC := g.listItemsAreRC(field.TypeExpr, name)
				g.writeln("%s = kl_list_new(%s);", target, g.boolToC(itemsRC))
				// Resolve element type from List<T>
				elemType := ""
				if gt, ok := field.TypeExpr.(*parser.GenericType); ok && gt.Name == "List" && len(gt.TypeArgs) > 0 {
					elemType = g.typeToC(gt.TypeArgs[0], name)
				}
				for _, elem := range arr.Elements {
					valStr := g.exprToC(elem)
					if sl, ok := elem.(*parser.StructLit); ok && elemType != "" {
						valStr = g.emitTypedStructLit(sl, elemType)
					}
					g.writeln("kl_list_push(%s, %s);", target, g.listPushCast(valStr, elemType))
				}
			} else {
				g.writeln("%s = %s;", target, g.exprToC(field.Value))
			}
		}
	}

	if cls.Constructor != nil && cls.Constructor.Body != nil {
		g.withClass(name, cls, func() {
			g.localVars = map[string]string{}
			g.localVarTypes = map[string]parser.TypeExpr{}
			// Register constructor params as local vars so they shadow fields
			for _, p := range cls.Constructor.Params {
				g.localVars[p.Name] = g.typeToC(p.TypeExpr, name)
			}
			g.emitBlock(cls.Constructor.Body, name)
			g.localVars = nil
			g.localVarTypes = nil
		})
	}

	// In DLL mode, initialize function pointers for hot-reloadable methods
	if g.DLLMode {
		for _, m := range cls.Methods {
			if m.Name == "main" {
				continue
			}
			g.writeln("self->_fn_%s = %s_%s;", m.Name, name, m.Name)
		}
	}

	g.writeln("return self;")
	g.indent--
	g.writeln("}")
	g.writeln("")
}

func (g *Generator) emitMethod(className string, method *parser.MethodDecl) {
	retType := g.returnTypeToC(method.ReturnType, className)
	params := g.methodParams(className, method)
	g.writeln("%s %s_%s(%s) {", retType, className, method.Name, params)
	g.indent++

	g.localVars = map[string]string{}
	g.localVarTypes = map[string]parser.TypeExpr{}
	g.pushScope()
	// Register params as local vars
	for _, p := range method.Params {
		g.localVars[p.Name] = g.typeToC(p.TypeExpr, className)
	}

	// In DLL mode, track when we're inside main() for function pointer indirection
	if g.DLLMode && method.Name == "main" {
		g.dllInsideMain = true
	}

	if method.Body != nil {
		g.emitBlock(method.Body, className)
	}

	if g.DLLMode && method.Name == "main" {
		g.dllInsideMain = false
	}

	// Emit scope cleanup at end of method (for non-void methods, return handles it)
	if retType == "void" {
		g.emitScopeCleanup("")
	}
	g.popScope()
	g.localVars = nil
	g.localVarTypes = nil

	g.indent--
	g.writeln("}")
	g.writeln("")
}

func (g *Generator) emitGenericInstantiation(className string, method *parser.MethodDecl, typeMap map[string]string) {
	g.typeSubstitutions = typeMap
	defer func() { g.typeSubstitutions = nil }()

	mangledName := g.mangledGenericName(className, method.Name, typeMap, method.TypeParams)
	retType := g.returnTypeToC(method.ReturnType, className)
	params := g.methodParams(className, method)
	g.writeln("%s %s(%s) {", retType, mangledName, params)
	g.indent++

	g.localVars = map[string]string{}
	g.localVarTypes = map[string]parser.TypeExpr{}
	for _, p := range method.Params {
		g.localVars[p.Name] = g.typeToC(p.TypeExpr, className)
	}

	if method.Body != nil {
		g.emitBlock(method.Body, className)
	}

	g.localVars = nil
	g.localVarTypes = nil
	g.indent--
	g.writeln("}")
	g.writeln("")
}

func (g *Generator) emitGenericClassCode() {
	for gcName, cls := range g.genericClasses {
		instances := g.genericClassInstances[gcName]
		if len(instances) == 0 {
			continue
		}

		for _, typeMap := range instances {
			mangledClass := gcName + "_" + g.mangledSuffix(typeMap, cls.TypeParams)

			// Register in classes map so member access resolution works
			g.classes[mangledClass] = cls

			// Forward declaration
			g.writeln("typedef struct %s %s;", mangledClass, mangledClass)
		}
	}
	g.writeln("")

	for gcName, cls := range g.genericClasses {
		instances := g.genericClassInstances[gcName]
		for _, typeMap := range instances {
			mangledClass := gcName + "_" + g.mangledSuffix(typeMap, cls.TypeParams)
			g.typeSubstitutions = typeMap

			// Struct definition
			g.writeln("struct %s {", mangledClass)
			g.indent++
			g.writeln("KlHeader _header;")
			for _, field := range cls.Fields {
				cType := g.typeToC(field.TypeExpr, mangledClass)
				if field.Inferred {
					cType = g.inferCType(field.Value)
				}
				g.writeln("%s %s;", cType, field.Name)
			}
			g.indent--
			g.writeln("};")
			g.writeln("")

			g.typeSubstitutions = nil
		}
	}

	// Type IDs for generic class instances (continue from existing IDs)
	// Count existing non-generic type IDs
	hasGenericInstances := false
	for _, instances := range g.genericClassInstances {
		if len(instances) > 0 {
			hasGenericInstances = true
			break
		}
	}
	if hasGenericInstances {
		// Find max existing type ID
		maxId := g.countTypeIds()
		g.writeln("enum {")
		g.indent++
		for gcName, cls := range g.genericClasses {
			for _, typeMap := range g.genericClassInstances[gcName] {
				mangledClass := gcName + "_" + g.mangledSuffix(typeMap, cls.TypeParams)
				maxId++
				g.writeln("KLTYPE_%s = %d,", mangledClass, maxId)
			}
		}
		g.indent--
		g.writeln("};")
		g.writeln("")
	}

	// Prototypes
	for gcName, cls := range g.genericClasses {
		for _, typeMap := range g.genericClassInstances[gcName] {
			mangledClass := gcName + "_" + g.mangledSuffix(typeMap, cls.TypeParams)
			g.typeSubstitutions = typeMap

			// Constructor prototype
			params := g.constructorParams(cls, mangledClass)
			g.writeln("%s* %s_new(%s);", mangledClass, mangledClass, params)

			// Method prototypes
			for _, method := range cls.Methods {
				retType := g.returnTypeToC(method.ReturnType, mangledClass)
				mParams := g.genericClassMethodParams(mangledClass, method)
				g.writeln("%s %s_%s(%s);", retType, mangledClass, method.Name, mParams)
			}

			g.typeSubstitutions = nil
		}
	}
	g.writeln("")

	// Implementations
	for gcName, cls := range g.genericClasses {
		for _, typeMap := range g.genericClassInstances[gcName] {
			mangledClass := gcName + "_" + g.mangledSuffix(typeMap, cls.TypeParams)
			g.typeSubstitutions = typeMap

			// Emit simple destructor/tracer for generic class instance
			g.writeln("static void %s_destroy(KlObject* _obj) {", mangledClass)
			g.indent++
			g.writeln("%s* self = (%s*)_obj;", mangledClass, mangledClass)
			for _, field := range cls.Fields {
				cType := g.typeToC(field.TypeExpr, mangledClass)
				if field.Inferred {
					cType = g.inferCType(field.Value)
				}
				if g.isRefCountedType(cType) {
					g.writeln("kl_release(self->%s);", field.Name)
				}
			}
			g.indent--
			g.writeln("}")
			g.writeln("")
			g.writeln("static void %s_trace(KlObject* _obj, void (*visit)(KlObject*)) {", mangledClass)
			g.indent++
			g.writeln("%s* self = (%s*)_obj;", mangledClass, mangledClass)
			for _, field := range cls.Fields {
				cType := g.typeToC(field.TypeExpr, mangledClass)
				if field.Inferred {
					cType = g.inferCType(field.Value)
				}
				if g.isRefCountedType(cType) {
					g.writeln("if (self->%s) visit((KlObject*)self->%s);", field.Name, field.Name)
				}
			}
			g.indent--
			g.writeln("}")
			g.writeln("")

			// Constructor
			params := g.constructorParams(cls, mangledClass)
			g.writeln("%s* %s_new(%s) {", mangledClass, mangledClass, params)
			g.indent++
			g.writeln("%s* self = (%s*)kl_alloc_rc(sizeof(%s), %s_destroy);", mangledClass, mangledClass, mangledClass, mangledClass)
			g.writeln("self->_header.type_id = KLTYPE_%s;", mangledClass)
			g.writeln("self->_header.tracer = %s_trace;", mangledClass)

			// Field defaults
			for _, field := range cls.Fields {
				if field.Value != nil {
					g.writeln("self->%s = %s;", field.Name, g.exprToC(field.Value))
				}
			}

			// Constructor body
			if cls.Constructor != nil && cls.Constructor.Body != nil {
				g.currentClassName = mangledClass
				g.currentClass = cls
				g.localVars = map[string]string{}
				g.localVarTypes = map[string]parser.TypeExpr{}
				for _, p := range cls.Constructor.Params {
					g.localVars[p.Name] = g.typeToC(p.TypeExpr, mangledClass)
				}
				g.emitBlock(cls.Constructor.Body, mangledClass)
				g.localVars = nil
				g.localVarTypes = nil
			}

			g.writeln("return self;")
			g.indent--
			g.writeln("}")
			g.writeln("")

			// Methods
			for _, method := range cls.Methods {
				retType := g.returnTypeToC(method.ReturnType, mangledClass)
				mParams := g.genericClassMethodParams(mangledClass, method)
				g.writeln("%s %s_%s(%s) {", retType, mangledClass, method.Name, mParams)
				g.indent++

				g.currentClassName = mangledClass
				g.currentClass = cls
				g.localVars = map[string]string{}
				g.localVarTypes = map[string]parser.TypeExpr{}
				for _, p := range method.Params {
					g.localVars[p.Name] = g.typeToC(p.TypeExpr, mangledClass)
				}
				if method.Body != nil {
					g.emitBlock(method.Body, mangledClass)
				}
				g.localVars = nil
				g.localVarTypes = nil

				g.indent--
				g.writeln("}")
				g.writeln("")
			}

			g.typeSubstitutions = nil
		}
	}
}

func (g *Generator) countTypeIds() int {
	count := 0
	for _, cls := range g.allClasses() {
		count += g.countTypeIdsInClass(cls)
	}
	return count
}

func (g *Generator) countTypeIdsInClass(cls *parser.ClassDecl) int {
	if len(cls.TypeParams) > 0 {
		return 0
	}
	count := 1
	for _, nested := range cls.Classes {
		count += g.countTypeIdsInClass(nested)
	}
	return count
}

func (g *Generator) genericClassMethodParams(className string, method *parser.MethodDecl) string {
	parts := []string{fmt.Sprintf("%s* self", className)}
	for _, p := range method.Params {
		parts = append(parts, fmt.Sprintf("%s %s", g.typeToC(p.TypeExpr, className), p.Name))
	}
	return strings.Join(parts, ", ")
}

func (g *Generator) emitProperty(className string, prop *parser.PropertyDecl) {
	propType := g.typeToC(prop.TypeExpr, className)

	if prop.Getter != nil {
		g.writeln("%s %s_get_%s(%s* self) {", propType, className, prop.Name, className)
		g.indent++
		g.localVars = map[string]string{}
		g.localVarTypes = map[string]parser.TypeExpr{}
		g.writeln("return %s;", g.exprToC(prop.Getter))
		g.localVars = nil
		g.localVarTypes = nil
		g.indent--
		g.writeln("}")
		g.writeln("")
	}

	if prop.Setter != nil {
		g.writeln("void %s_set_%s(%s* self, %s value) {", className, prop.Name, className, propType)
		g.indent++
		g.localVars = map[string]string{"value": ""}
		g.localVarTypes = map[string]parser.TypeExpr{}
		g.emitBlock(prop.Setter, className)
		g.localVars = nil
		g.localVarTypes = nil
		g.indent--
		g.writeln("}")
		g.writeln("")
	}
}

// --- Statement emission ---

func (g *Generator) emitBlock(block *parser.Block, className string) {
	for _, stmt := range block.Stmts {
		g.emitStmt(stmt, className)
	}
}

func (g *Generator) emitStmt(stmt parser.Stmt, className string) {
	switch s := stmt.(type) {
	case *parser.VarDecl:
		cType := g.inferCType(s.Value)
		if s.TypeExpr != nil {
			cType = g.typeToC(s.TypeExpr, className)
		}
		if g.localVars != nil {
			g.localVars[s.Name] = cType
		}
		if g.localVarTypes != nil && s.TypeExpr != nil {
			g.localVarTypes[s.Name] = s.TypeExpr
		}
		// Track RC locals for scope-exit cleanup
		if g.isRefCountedType(cType) {
			g.pushScopeVar(s.Name, cType)
		}
		if arr, ok := s.Value.(*parser.ArrayLit); ok {
			// Array literal: expand to list_new + push calls
			itemsRC := g.listItemsAreRC(s.TypeExpr, className)
			g.writeln("%s %s = kl_list_new(%s);", cType, s.Name, g.boolToC(itemsRC))
			elemType := ""
			if s.TypeExpr != nil {
				if gt, ok := s.TypeExpr.(*parser.GenericType); ok && gt.Name == "List" && len(gt.TypeArgs) > 0 {
					elemType = g.typeToC(gt.TypeArgs[0], className)
				}
			}
			for _, elem := range arr.Elements {
				g.writeln("kl_list_push(%s, %s);", s.Name, g.listPushCast(g.exprToC(elem), elemType))
			}
		} else if s.Value != nil {
			valStr := g.exprToC(s.Value)
			// Struct literals need the type for compound literal syntax
			if sl, ok := s.Value.(*parser.StructLit); ok && s.TypeExpr != nil {
				valStr = g.emitTypedStructLit(sl, cType)
			}
			g.writeln("%s %s = %s;", cType, s.Name, valStr)
		} else {
			g.writeln("%s %s;", cType, s.Name)
		}

	case *parser.AssignStmt:
		// Direct index assignment: list[i] = val → kl_list_set(list, i, val)
		if idx, ok := s.Target.(*parser.IndexExpr); ok && s.Op == "=" {
			obj := g.exprToC(idx.Object)
			idxStr := g.exprToC(idx.Index)
			valueStr := g.exprToC(s.Value)
			elemType := g.resolveForElemType(idx.Object)
			g.writeln("kl_list_set(%s, %s, %s);", obj, idxStr, g.listPushCast(valueStr, elemType))
			break
		}
		targetStr := g.exprToC(s.Target)
		valueStr := g.exprToC(s.Value)
		targetCType := g.resolveAssignTargetType(s.Target, className)
		if s.Op == "=" && g.isRefCountedType(targetCType) {
			isWeak := g.isWeakAssignTarget(s.Target, className)
			if isWeak {
				g.writeln("kl_weak_assign((KlWeakSlot**)&%s, %s);", targetStr, valueStr)
			} else {
				g.writeln("kl_strong_assign((void**)&%s, %s);", targetStr, valueStr)
			}
		} else if s.Op != "=" && g.isVecType(g.resolveExprType(s.Target)) {
			// Compound assignment on vec types: a += b → a = vec_add(a, b)
			vecType := g.resolveExprType(s.Target)
			var opFunc string
			switch s.Op {
			case "+=":
				opFunc = vecType + "_add"
			case "-=":
				opFunc = vecType + "_sub"
			case "*=":
				opFunc = vecType + "_scale"
			}
			if opFunc != "" {
				g.writeln("%s = %s(%s, %s);", targetStr, opFunc, targetStr, valueStr)
			} else {
				g.writeln("%s %s %s;", targetStr, s.Op, valueStr)
			}
		} else {
			g.writeln("%s %s %s;", targetStr, s.Op, valueStr)
		}

	case *parser.ReturnStmt:
		if s.Value != nil {
			valStr := g.exprToC(s.Value)
			// Determine if we're returning a local var (ownership transfer)
			returnedVar := g.identifyReturnedLocalVar(s.Value)
			// Emit cleanup for all scope vars except the returned one
			g.emitScopeCleanup(returnedVar)
			g.writeln("return %s;", valStr)
		} else {
			g.emitScopeCleanup("")
			g.writeln("return;")
		}

	case *parser.IfStmt:
		g.emitIf(s, className)

	case *parser.ForStmt:
		g.emitFor(s, className)

	case *parser.WhileStmt:
		g.writeln("while (%s) {", g.exprToC(s.Condition))
		g.indent++
		g.pushScope()
		g.emitBlock(s.Body, className)
		g.emitScopeCleanupCurrentOnly()
		g.popScope()
		g.indent--
		g.writeln("}")

	case *parser.ExprStmt:
		g.writeln("%s;", g.exprToC(s.Expr))

	case *parser.WithStmt:
		g.withModules = append(g.withModules, s.Module)
		g.emitBlock(s.Body, className)
		g.withModules = g.withModules[:len(g.withModules)-1]

	case *parser.InlineCStmt:
		g.writeln("%s", s.Code)

	case *parser.Block:
		g.emitBlock(s, className)
	}
}

func (g *Generator) emitIf(s *parser.IfStmt, className string) {
	// Type narrowing for "if expr is Type" patterns
	var narrowVarName string
	var narrowOldType string
	var narrowIsField bool
	if isExpr, ok := s.Condition.(*parser.IsExpr); ok {
		if ident, ok := isExpr.Expr.(*parser.Ident); ok {
			narrowVarName = ident.Name
			if g.localVars != nil {
				narrowOldType = g.localVars[ident.Name]
				// Check if this is a field (not already a local var)
				if _, isLocal := g.localVars[ident.Name]; !isLocal {
					narrowIsField = true
				}
			}
		}
		_ = narrowIsField
	}

	g.writeln("if (%s) {", g.exprToC(s.Condition))
	g.indent++

	// Apply type narrowing in then block
	if narrowVarName != "" {
		if isExpr, ok := s.Condition.(*parser.IsExpr); ok {
			fullType := g.resolveFullClassName(isExpr.TypeName, className)
			if g.localVars != nil {
				// For fields (including weak fields), create a local variable
				// to cache the field read with the narrowed type
				if narrowIsField {
					fieldExpr := g.resolveIdent(narrowVarName)
					g.writeln("%s* %s = (%s*)%s;", fullType, narrowVarName, fullType, fieldExpr)
				}
				g.localVars[narrowVarName] = fullType + "*"
			}
		}
	}

	if s.Then != nil {
		g.emitBlock(s.Then, className)
	}
	if s.ThenStmt != nil {
		g.emitStmt(s.ThenStmt, className)
	}
	g.indent--

	// Restore type after then block
	if narrowVarName != "" && g.localVars != nil {
		g.localVars[narrowVarName] = narrowOldType
	}

	if s.Else != nil {
		switch e := s.Else.(type) {
		case *parser.IfStmt:
			g.write("} else ")
			g.emitIf(e, className)
			return
		case *parser.Block:
			g.writeln("} else {")
			g.indent++
			g.emitBlock(e, className)
			g.indent--
		}
	}
	g.writeln("}")
}

func (g *Generator) emitFor(s *parser.ForStmt, className string) {
	// Check if iterating over a numeric range: for i in 10 → for (int i = 0; i < 10; i++)
	if g.isNumericExpr(s.Iterable) {
		limit := g.exprToC(s.Iterable)
		if g.localVars != nil {
			g.localVars[s.VarName] = "int"
		}
		g.writeln("for (int %s = 0; %s < %s; %s++) {", s.VarName, s.VarName, limit, s.VarName)
		g.indent++
		g.pushScope()
		g.emitBlock(s.Body, className)
		g.emitScopeCleanupCurrentOnly()
		g.popScope()
		g.indent--
		g.writeln("}")
		return
	}

	// List iteration: for item in list → for (int _i = 0; _i < list->count; _i++)
	iter := g.exprToC(s.Iterable)
	elemType := g.resolveForElemType(s.Iterable)

	if g.localVars != nil {
		g.localVars[s.VarName] = elemType
	}

	g.writeln("for (int _i = 0; _i < %s->count; _i++) {", iter)
	g.indent++
	g.pushScope()
	g.writeln("%s %s = %s;", elemType, s.VarName, g.listGetCast(elemType, iter, "_i"))
	g.emitBlock(s.Body, className)
	g.emitScopeCleanupCurrentOnly()
	g.popScope()
	g.indent--
	g.writeln("}")
}

func (g *Generator) isNumericExpr(expr parser.Expr) bool {
	switch expr.(type) {
	case *parser.IntLit:
		return true
	case *parser.FloatLit:
		return true
	case *parser.BinaryExpr:
		e := expr.(*parser.BinaryExpr)
		return g.isNumericExpr(e.Left) && g.isNumericExpr(e.Right)
	case *parser.Ident:
		// Check if it's a known int/float local var
		if g.localVars != nil {
			if t, ok := g.localVars[expr.(*parser.Ident).Name]; ok {
				return t == "int" || t == "float"
			}
		}
		return false
	}
	return false
}

func (g *Generator) isPrimitiveType(cType string) bool {
	return cType == "int" || cType == "float" || cType == "bool"
}

func (g *Generator) isValueType(cType string) bool {
	switch cType {
	case "vec2", "vec3", "vec4", "mat4", "quat",
		"Color", "Rectangle", "Camera2D", "Camera3D",
		"Texture2D", "Font", "Sound", "Music",
		"KlRandom":
		return true
	}
	return false
}

// isVecType returns true for vector types that support arithmetic operators
func (g *Generator) isVecType(cType string) bool {
	switch cType {
	case "vec2", "vec3", "vec4", "quat":
		return true
	}
	return false
}

// tryVecBinaryOp handles binary operators on vec types, returning the function call string
// or "" if this isn't a vec operation.
func (g *Generator) tryVecBinaryOp(e *parser.BinaryExpr) string {
	leftType := g.resolveExprType(e.Left)
	rightType := g.resolveExprType(e.Right)

	// Determine the vec type (prefer left side)
	vecType := ""
	if g.isVecType(leftType) {
		vecType = leftType
	} else if g.isVecType(rightType) {
		vecType = rightType
	}
	if vecType == "" {
		return ""
	}

	leftStr := g.exprToC(e.Left)
	rightStr := g.exprToC(e.Right)

	switch e.Op {
	case "+":
		return fmt.Sprintf("%s_add(%s, %s)", vecType, leftStr, rightStr)
	case "-":
		return fmt.Sprintf("%s_sub(%s, %s)", vecType, leftStr, rightStr)
	case "*":
		// Component-wise multiply if both are vecs, scale if one is scalar
		if g.isVecType(leftType) && g.isVecType(rightType) {
			return fmt.Sprintf("%s_mul(%s, %s)", vecType, leftStr, rightStr)
		}
		// One side is scalar → use scale
		if g.isVecType(leftType) {
			return fmt.Sprintf("%s_scale(%s, %s)", vecType, leftStr, rightStr)
		}
		return fmt.Sprintf("%s_scale(%s, %s)", vecType, rightStr, leftStr)
	case "/":
		// vec / scalar → scale by 1/s
		if g.isVecType(leftType) && !g.isVecType(rightType) {
			return fmt.Sprintf("%s_scale(%s, 1.0f / (%s))", vecType, leftStr, rightStr)
		}
	}
	return ""
}

// inferVectorFuncReturnType returns the C type for known vector/math global functions
func (g *Generator) inferVectorFuncReturnType(name string) string {
	// Functions that return scalars
	scalarFuncs := map[string]bool{
		"vec2_dot": true, "vec3_dot": true, "vec4_dot": true,
		"vec2_length": true, "vec3_length": true, "vec4_length": true,
		"vec3_length_sq": true, "vec3_distance": true,
		"quat_length": true,
	}
	if scalarFuncs[name] {
		return "float"
	}
	// Functions that return their type prefix
	prefixes := []struct{ prefix, typ string }{
		{"vec2_", "vec2"}, {"vec3_", "vec3"}, {"vec4_", "vec4"},
		{"mat4_", "mat4"}, {"quat_", "quat"},
	}
	for _, p := range prefixes {
		if strings.HasPrefix(name, p.prefix) {
			return p.typ
		}
	}
	return ""
}

// inferRlReturnType returns the C type for raylib wrapper function return values
func (g *Generator) inferRlReturnType(funcName string) string {
	switch funcName {
	// bool returns
	case "window_should_close", "is_window_resized",
		"is_key_pressed", "is_key_down", "is_key_released", "is_key_up",
		"is_mouse_button_pressed", "is_mouse_button_down", "is_mouse_button_released",
		"is_gamepad_available", "is_gamepad_button_pressed", "is_gamepad_button_down":
		return "bool"
	// int returns
	case "get_screen_width", "get_screen_height", "get_fps", "measure_text",
		"get_mouse_x", "get_mouse_y":
		return "int"
	// float returns
	case "get_frame_time", "get_time", "get_mouse_wheel_move", "get_gamepad_axis_movement":
		return "float"
	// vec2 returns
	case "get_mouse_position", "get_screen_size":
		return "vec2"
	// Texture2D returns
	case "load_texture":
		return "Texture2D"
	// Font returns
	case "load_font":
		return "Font"
	// Sound returns
	case "load_sound":
		return "Sound"
	// Music returns
	case "load_music_stream":
		return "Music"
	// Color returns
	case "color", "color_rgb":
		return "Color"
	// Rectangle returns
	case "rect":
		return "Rectangle"
	// Camera returns
	case "camera2d":
		return "Camera2D"
	case "camera3d":
		return "Camera3D"
	}
	// void by default (drawing functions etc.)
	return "int"
}

// resolveExprType returns the C type string for an expression
func (g *Generator) resolveExprType(expr parser.Expr) string {
	if ident, ok := expr.(*parser.Ident); ok {
		if g.localVars != nil {
			if t, ok := g.localVars[ident.Name]; ok && t != "" {
				return t
			}
		}
		if g.currentClass != nil {
			for _, f := range g.currentClass.Fields {
				if f.Name == ident.Name {
					return g.fieldCType(f, g.currentClassName)
				}
			}
		}
	}
	if call, ok := expr.(*parser.CallExpr); ok {
		if ident, ok := call.Callee.(*parser.Ident); ok {
			switch ident.Name {
			case "vec2":
				return "vec2"
			case "vec3":
				return "vec3"
			case "vec4":
				return "vec4"
			case "quat":
				return "quat"
			case "Color":
				return "Color"
			case "Rectangle":
				return "Rectangle"
			case "Random":
				return "KlRandom"
			}
		}
		// Check rl module function calls for return type
		if member, ok := call.Callee.(*parser.MemberExpr); ok {
			if ident, ok := member.Object.(*parser.Ident); ok {
				if ident.Name == "rl" {
					return g.inferRlReturnType(member.Field)
				}
			}
		}
	}
	// Handle index access: list[i] → resolve element type
	if idx, ok := expr.(*parser.IndexExpr); ok {
		return g.resolveForElemType(idx.Object)
	}
	// Handle chained member access: ball.position → resolve field type on class
	if member, ok := expr.(*parser.MemberExpr); ok {
		objType := g.resolveExprType(member.Object)
		if objType != "" {
			// Strip pointer suffix to get class name
			className := strings.TrimSuffix(objType, "*")
			if cls, ok := g.classes[className]; ok {
				for _, f := range cls.Fields {
					if f.Name == member.Field {
						return g.fieldCType(f, className)
					}
				}
				// Check parent class fields
				if cls.Parent != "" {
					prefix := g.extractPrefix(className, cls.Name)
					parentName := g.resolveParentName(prefix, cls.Parent)
					if parent, ok := g.classes[parentName]; ok {
						for _, f := range parent.Fields {
							if f.Name == member.Field {
								return g.fieldCType(f, parentName)
							}
						}
					}
				}
			}
			// If the object is a value type, resolve common field types
			if g.isValueType(objType) {
				switch member.Field {
				case "x", "y", "z", "w":
					return "float"
				case "r", "g", "b", "a":
					return "int"
				}
			}
		}
	}
	return ""
}

// Wrap a value for kl_list_push: primitives need (void*)(intptr_t) cast
func (g *Generator) listPushCast(valStr string, cType string) string {
	if g.isPrimitiveType(cType) {
		return fmt.Sprintf("(void*)(intptr_t)(%s)", valStr)
	}
	return valStr
}

// Unwrap a value from kl_list_get: primitives need (type)(intptr_t) cast
func (g *Generator) emitIndexGet(e *parser.IndexExpr) string {
	obj := g.exprToC(e.Object)
	idx := g.exprToC(e.Index)
	elemType := g.resolveForElemType(e.Object)
	return g.listGetCast(elemType, obj, idx)
}

func (g *Generator) listGetCast(elemType string, listVar string, indexVar string) string {
	if g.isPrimitiveType(elemType) {
		return fmt.Sprintf("((%s)(intptr_t)kl_list_get(%s, %s))", elemType, listVar, indexVar)
	}
	return fmt.Sprintf("((%s)kl_list_get(%s, %s))", elemType, listVar, indexVar)
}

// listGetCastDirect casts a void* expression to the element type
func (g *Generator) listGetCastDirect(elemType string, expr string) string {
	if g.isPrimitiveType(elemType) {
		return fmt.Sprintf("((%s)(intptr_t)%s)", elemType, expr)
	}
	return fmt.Sprintf("((%s)%s)", elemType, expr)
}

// emitListLambdaMethod handles list methods that take lambda arguments.
// These are handled before general arg emission to control lambda param types.
func (g *Generator) emitListLambdaMethod(e *parser.CallExpr, member *parser.MemberExpr) string {
	obj := g.exprToC(member.Object)
	elemType := g.resolveForElemType(member.Object)
	methodName := member.Field

	// Extract the lambda expression from the first argument
	lambda, ok := e.Args[0].(*parser.LambdaExpr)
	if !ok {
		argStr := g.exprToC(e.Args[0])
		return fmt.Sprintf("/* %s: expected lambda */ %s", methodName, argStr)
	}

	// Helper: emit element variable declaration from list->data[indexVar]
	emitElemDecl := func(paramName string, indexVar string) {
		if g.isPrimitiveType(elemType) {
			g.writeln("%s %s = (%s)(intptr_t)%s->data[%s];", elemType, paramName, elemType, obj, indexVar)
		} else {
			g.writeln("%s %s = (%s)%s->data[%s];", elemType, paramName, elemType, obj, indexVar)
		}
	}

	// Helper: temporarily register lambda params as local vars, return restore func
	withLambdaParams := func(paramTypes []string) func() {
		saved := make([]struct {
			name string
			had  bool
			val  string
		}, len(lambda.Params))
		for i, p := range lambda.Params {
			if i < len(paramTypes) {
				saved[i].name = p.Name
				saved[i].val, saved[i].had = g.localVars[p.Name]
				g.localVars[p.Name] = paramTypes[i]
			}
		}
		return func() {
			for _, s := range saved {
				if s.name == "" {
					continue
				}
				if s.had {
					g.localVars[s.name] = s.val
				} else {
					delete(g.localVars, s.name)
				}
			}
		}
	}

	// Helper: get the lambda body expression (assumes single return statement)
	getLambdaExpr := func() parser.Expr {
		if len(lambda.Body.Stmts) > 0 {
			if ret, ok := lambda.Body.Stmts[0].(*parser.ReturnStmt); ok {
				return ret.Value
			}
		}
		return nil
	}

	id := g.lambdaCounter
	g.lambdaCounter++

	switch methodName {

	case "remove_all":
		restore := withLambdaParams([]string{elemType})
		dstVar := fmt.Sprintf("_dst_%d", id)
		g.writeln("{")
		g.indent++
		g.writeln("int %s = 0;", dstVar)
		g.writeln("for (int _i = 0; _i < %s->count; _i++) {", obj)
		g.indent++
		emitElemDecl(lambda.Params[0].Name, "_i")
		if expr := getLambdaExpr(); expr != nil {
			condStr := g.exprToC(expr)
			g.writeln("if (%s) { if (%s->items_are_rc && %s->data[_i]) kl_release(%s->data[_i]); }", condStr, obj, obj, obj)
			g.writeln("else { %s->data[%s++] = %s->data[_i]; }", obj, dstVar, obj)
		}
		g.indent--
		g.writeln("}")
		g.writeln("%s->count = %s;", obj, dstVar)
		g.indent--
		g.writeln("}")
		restore()
		return "(void)0"

	case "filter":
		restore := withLambdaParams([]string{elemType})
		resultVar := fmt.Sprintf("_filtered_%d", id)
		g.writeln("KlList* %s = kl_list_new(%s->items_are_rc);", resultVar, obj)
		g.writeln("for (int _i = 0; _i < %s->count; _i++) {", obj)
		g.indent++
		emitElemDecl(lambda.Params[0].Name, "_i")
		if expr := getLambdaExpr(); expr != nil {
			condStr := g.exprToC(expr)
			g.writeln("if (%s) kl_list_push(%s, %s->data[_i]);", condStr, resultVar, obj)
		}
		g.indent--
		g.writeln("}")
		restore()
		return resultVar

	case "map":
		restore := withLambdaParams([]string{elemType})
		resultVar := fmt.Sprintf("_mapped_%d", id)
		g.writeln("KlList* %s = kl_list_new(%s->items_are_rc);", resultVar, obj)
		g.writeln("for (int _i = 0; _i < %s->count; _i++) {", obj)
		g.indent++
		emitElemDecl(lambda.Params[0].Name, "_i")
		if expr := getLambdaExpr(); expr != nil {
			valStr := g.exprToC(expr)
			mappedType := g.resolveExprType(expr)
			if g.isPrimitiveType(mappedType) {
				g.writeln("kl_list_push(%s, (void*)(intptr_t)(%s));", resultVar, valStr)
			} else {
				g.writeln("kl_list_push(%s, %s);", resultVar, valStr)
			}
		}
		g.indent--
		g.writeln("}")
		restore()
		return resultVar

	case "find":
		restore := withLambdaParams([]string{elemType})
		resultVar := fmt.Sprintf("_found_%d", id)
		g.writeln("void* %s = NULL;", resultVar)
		g.writeln("for (int _i = 0; _i < %s->count; _i++) {", obj)
		g.indent++
		emitElemDecl(lambda.Params[0].Name, "_i")
		if expr := getLambdaExpr(); expr != nil {
			condStr := g.exprToC(expr)
			g.writeln("if (%s) { %s = %s->data[_i]; break; }", condStr, resultVar, obj)
		}
		g.indent--
		g.writeln("}")
		restore()
		return g.listGetCastDirect(elemType, resultVar)

	case "find_index":
		restore := withLambdaParams([]string{elemType})
		resultVar := fmt.Sprintf("_fidx_%d", id)
		g.writeln("int %s = -1;", resultVar)
		g.writeln("for (int _i = 0; _i < %s->count; _i++) {", obj)
		g.indent++
		emitElemDecl(lambda.Params[0].Name, "_i")
		if expr := getLambdaExpr(); expr != nil {
			condStr := g.exprToC(expr)
			g.writeln("if (%s) { %s = _i; break; }", condStr, resultVar)
		}
		g.indent--
		g.writeln("}")
		restore()
		return resultVar

	case "sort":
		if len(lambda.Params) < 2 {
			return fmt.Sprintf("/* sort: expected 2 lambda params */ (void)0")
		}
		cmpName := fmt.Sprintf("_list_cmp_%d", id)

		prevOut := g.out
		prevIndent := g.indent
		prevLocalVars := g.localVars
		prevLocalVarTypes := g.localVarTypes

		g.out = strings.Builder{}
		g.indent = 1
		g.localVars = map[string]string{}
		g.localVarTypes = map[string]parser.TypeExpr{}

		p1 := lambda.Params[0].Name
		p2 := lambda.Params[1].Name
		g.localVars[p1] = elemType
		g.localVars[p2] = elemType

		if g.isPrimitiveType(elemType) {
			g.writeln("%s %s = (%s)(intptr_t)*(void**)_a;", elemType, p1, elemType)
			g.writeln("%s %s = (%s)(intptr_t)*(void**)_b;", elemType, p2, elemType)
		} else {
			g.writeln("%s %s = (%s)*(void**)_a;", elemType, p1, elemType)
			g.writeln("%s %s = (%s)*(void**)_b;", elemType, p2, elemType)
		}

		if expr := getLambdaExpr(); expr != nil {
			exprStr := g.exprToC(expr)
			g.writeln("float _r = %s;", exprStr)
			g.writeln("return (_r > 0) ? 1 : (_r < 0) ? -1 : 0;")
		}

		body := g.out.String()
		g.out = prevOut
		g.indent = prevIndent
		g.localVars = prevLocalVars
		g.localVarTypes = prevLocalVarTypes

		fmt.Fprintf(&g.lambdaDefs, "static int %s(const void* _a, const void* _b) {\n%s}\n\n", cmpName, body)
		return fmt.Sprintf("qsort(%s->data, %s->count, sizeof(void*), %s)", obj, obj, cmpName)

	case "sort_by":
		cmpName := fmt.Sprintf("_list_cmp_%d", id)

		prevOut := g.out
		prevIndent := g.indent
		prevLocalVars := g.localVars
		prevLocalVarTypes := g.localVarTypes

		g.out = strings.Builder{}
		g.indent = 1
		g.localVars = map[string]string{}
		g.localVarTypes = map[string]parser.TypeExpr{}

		p1 := lambda.Params[0].Name
		g.localVars[p1] = elemType

		// Declare the param variable, then reuse it for both a and b
		g.writeln("%s %s;", elemType, p1)

		if expr := getLambdaExpr(); expr != nil {
			if g.isPrimitiveType(elemType) {
				g.writeln("%s = (%s)(intptr_t)*(void**)_a;", p1, elemType)
			} else {
				g.writeln("%s = (%s)*(void**)_a;", p1, elemType)
			}
			exprA := g.exprToC(expr)
			g.writeln("float _ka = %s;", exprA)

			if g.isPrimitiveType(elemType) {
				g.writeln("%s = (%s)(intptr_t)*(void**)_b;", p1, elemType)
			} else {
				g.writeln("%s = (%s)*(void**)_b;", p1, elemType)
			}
			exprB := g.exprToC(expr)
			g.writeln("float _kb = %s;", exprB)

			g.writeln("return (_ka > _kb) ? 1 : (_ka < _kb) ? -1 : 0;")
		}

		body := g.out.String()
		g.out = prevOut
		g.indent = prevIndent
		g.localVars = prevLocalVars
		g.localVarTypes = prevLocalVarTypes

		fmt.Fprintf(&g.lambdaDefs, "static int %s(const void* _a, const void* _b) {\n%s}\n\n", cmpName, body)
		return fmt.Sprintf("qsort(%s->data, %s->count, sizeof(void*), %s)", obj, obj, cmpName)
	}

	return fmt.Sprintf("/* unknown list method: %s */ (void)0", methodName)
}

func (g *Generator) resolveForElemType(iterable parser.Expr) string {
	if ident, ok := iterable.(*parser.Ident); ok {
		// Check local var type expressions for List<T>
		if g.localVarTypes != nil {
			if typeExpr, ok := g.localVarTypes[ident.Name]; ok {
				if gt, ok := typeExpr.(*parser.GenericType); ok && gt.Name == "List" && len(gt.TypeArgs) > 0 {
					return g.typeToC(gt.TypeArgs[0], g.currentClassName)
				}
			}
		}
		// Check field type on the current class
		if g.currentClass != nil {
			for _, f := range g.currentClass.Fields {
				if f.Name == ident.Name {
					if gt, ok := f.TypeExpr.(*parser.GenericType); ok && gt.Name == "List" && len(gt.TypeArgs) > 0 {
						return g.typeToC(gt.TypeArgs[0], g.currentClassName)
					}
				}
			}
		}
	}
	return "void*"
}

// --- Expression emission ---

func (g *Generator) exprToC(expr parser.Expr) string {
	switch e := expr.(type) {
	case *parser.IntLit:
		return e.Value
	case *parser.FloatLit:
		return e.Value + "f"
	case *parser.StringLit:
		return fmt.Sprintf("\"%s\"", e.Value)
	case *parser.BoolLit:
		if e.Value {
			return "true"
		}
		return "false"
	case *parser.Ident:
		return g.resolveIdent(e.Name)
	case *parser.ThisExpr:
		return "self"
	case *parser.BinaryExpr:
		// Check if operands are vec/value types that need function calls instead of C operators
		if vecResult := g.tryVecBinaryOp(e); vecResult != "" {
			return vecResult
		}
		return fmt.Sprintf("(%s %s %s)", g.exprToC(e.Left), e.Op, g.exprToC(e.Right))
	case *parser.UnaryExpr:
		if e.Op == "not" {
			return fmt.Sprintf("(!%s)", g.exprToC(e.Operand))
		}
		return fmt.Sprintf("(%s%s)", e.Op, g.exprToC(e.Operand))
	case *parser.CallExpr:
		return g.emitCall(e)
	case *parser.MemberExpr:
		// Check if this is a stdlib module constant: math.PI → KL_PI
		if ident, ok := e.Object.(*parser.Ident); ok {
			if consts, ok := stdlibModuleConstants[ident.Name]; ok {
				if cConst, ok := consts[e.Field]; ok {
					return cConst
				}
			}
			// Check constant namespaces: Colors.Red → RED, Key.Space → KEY_SPACE
			if consts, ok := stdlibConstNamespaces[ident.Name]; ok {
				if cConst, ok := consts[e.Field]; ok {
					return cConst
				}
			}
		}
		obj := g.exprToC(e.Object)
		if e.Optional {
			return fmt.Sprintf("(%s ? %s->%s : NULL)", obj, obj, e.Field)
		}
		// Value types use '.' instead of '->'
		objType := g.resolveExprType(e.Object)
		if g.isValueType(objType) {
			return fmt.Sprintf("%s.%s", obj, e.Field)
		}
		// Check if this is a method call without parens: ball.update → Main_Ball_update(ball)
		typeName := g.resolveExprTypeName(e.Object)
		if typeName != "" {
			if cls, ok := g.classes[typeName]; ok {
				// Check if field matches a method name (not a field)
				isField := false
				for _, f := range cls.Fields {
					if f.Name == e.Field {
						isField = true
						break
					}
				}
				if !isField {
					for _, m := range cls.Methods {
						if m.Name == e.Field {
							// Method call without parens
							return fmt.Sprintf("%s_%s(%s)", typeName, e.Field, obj)
						}
					}
				}
				// Check if we need _base access for inherited fields
				if cls.Parent != "" {
					isDirect := false
					for _, f := range cls.Fields {
						if f.Name == e.Field {
							isDirect = true
							break
						}
					}
					if !isDirect {
						prefix := g.extractPrefix(typeName, cls.Name)
						parentName := g.resolveParentName(prefix, cls.Parent)
						if parent, ok := g.classes[parentName]; ok {
							if g.findFieldInClass(parent, e.Field) {
								return fmt.Sprintf("%s->_base.%s", obj, e.Field)
							}
						}
					}
				}
			}
		}
		return fmt.Sprintf("%s->%s", obj, e.Field)
	case *parser.IndexExpr:
		return g.emitIndexGet(e)
	case *parser.IsExpr:
		fullTypeName := g.resolveFullClassName(e.TypeName, g.currentClassName)
		return fmt.Sprintf("(%s->_header.type_id == KLTYPE_%s)", g.exprToC(e.Expr), fullTypeName)
	case *parser.StructLit:
		return g.emitStructLit(e)
	case *parser.ArrayLit:
		return g.emitArrayLit(e)
	case *parser.NullCoalesce:
		left := g.exprToC(e.Left)
		right := g.exprToC(e.Right)
		return fmt.Sprintf("(%s ? %s : %s)", left, left, right)
	case *parser.InterpString:
		if len(e.Parts) > 0 {
			if sl, ok := e.Parts[0].(*parser.StringLit); ok {
				return fmt.Sprintf("\"%s\"", sl.Value)
			}
		}
		return "\"<interp>\""
	case *parser.LambdaExpr:
		return g.emitLambda(e)
	case *parser.SpreadExpr:
		return "/* ... */"
	}
	return "/* unknown expr */"
}

func (g *Generator) resolveIdent(name string) string {
	// Local variables take priority
	if g.localVars != nil {
		if _, ok := g.localVars[name]; ok {
			return name
		}
	}
	if g.currentClass != nil {
		// Direct field on this class
		for _, f := range g.currentClass.Fields {
			if f.Name == name {
				if g.isWeakField(g.currentClassName, name) {
					cType := g.fieldCType(f, g.currentClassName)
					targetType := strings.TrimSuffix(cType, "*")
					return fmt.Sprintf("((%s*)kl_weak_read(self->%s))", targetType, name)
				}
				return "self->" + name
			}
		}
		// Inherited field via _base
		if g.currentClass.Parent != "" {
			prefix := g.extractPrefix(g.currentClassName, g.currentClass.Name)
			parentName := g.resolveParentName(prefix, g.currentClass.Parent)
			if parent, ok := g.classes[parentName]; ok {
				if g.findFieldInClass(parent, name) {
					return "self->_base." + name
				}
			}
		}
	}
	// Check if it's a property — call the getter
	if g.isProperty(name) {
		return fmt.Sprintf("%s_get_%s(self)", g.currentClassName, name)
	}
	// Check active "with" modules for constants
	for i := len(g.withModules) - 1; i >= 0; i-- {
		mod := g.withModules[i]
		if consts, ok := stdlibModuleConstants[mod]; ok {
			if cConst, ok := consts[name]; ok {
				return cConst
			}
		}
	}
	return name
}

func (g *Generator) resolveExprTypeName(expr parser.Expr) string {
	if ident, ok := expr.(*parser.Ident); ok {
		// Check local vars for type info
		if g.localVars != nil {
			if t, ok := g.localVars[ident.Name]; ok && t != "" {
				return strings.TrimSuffix(t, "*")
			}
		}
		// Check fields on current class
		if g.currentClass != nil {
			for _, f := range g.currentClass.Fields {
				if f.Name == ident.Name && f.TypeExpr != nil {
					cType := g.typeToC(f.TypeExpr, g.currentClassName)
					return strings.TrimSuffix(cType, "*")
				}
			}
		}
	}
	// Index access: list[i] → element type name
	if idx, ok := expr.(*parser.IndexExpr); ok {
		elemType := g.resolveForElemType(idx.Object)
		return strings.TrimSuffix(elemType, "*")
	}
	return ""
}

func (g *Generator) emitCall(e *parser.CallExpr) string {
	// Handle list lambda methods before general arg emission (lambdas need element type hints)
	if member, ok := e.Callee.(*parser.MemberExpr); ok && g.isListExpr(member.Object) {
		switch member.Field {
		case "sort", "sort_by", "remove_all", "filter", "map", "find", "find_index":
			return g.emitListLambdaMethod(e, member)
		}
	}

	args := make([]string, len(e.Args))
	for i, arg := range e.Args {
		args[i] = g.exprToC(arg)
	}
	argStr := strings.Join(args, ", ")

	// Built-in: print with multiple args → kl_print_multi(...)
	if ident, ok := e.Callee.(*parser.Ident); ok && ident.Name == "print" && len(e.Args) > 1 {
		parts := make([]string, len(e.Args))
		for i, arg := range e.Args {
			parts[i] = fmt.Sprintf("kl_print_inline(%s)", g.exprToC(arg))
		}
		return strings.Join(parts, "; ") + "; kl_print_nl()"
	}

	// Built-in: wait(seconds) → kl_wait(seconds)
	if ident, ok := e.Callee.(*parser.Ident); ok && ident.Name == "wait" && len(e.Args) == 1 {
		return fmt.Sprintf("kl_wait(%s)", args[0])
	}

	// Check if callee is a generic method call
	if ident, ok := e.Callee.(*parser.Ident); ok {
		key := g.currentClassName + "_" + ident.Name
		if info, ok := g.genericMethods[key]; ok {
			typeMap := g.resolveGenericTypeArgs(info, e, g.currentClassName)
			mangledName := g.mangledGenericName(g.currentClassName, ident.Name, typeMap, info.method.TypeParams)
			if argStr != "" {
				return fmt.Sprintf("%s(self, %s)", mangledName, argStr)
			}
			return fmt.Sprintf("%s(self)", mangledName)
		}
	}

	// Check if callee is a generic class constructor call
	if ident, ok := e.Callee.(*parser.Ident); ok {
		gcName := g.resolveGenericClassName(ident.Name, g.currentClassName)
		if gcls, ok := g.genericClasses[gcName]; ok {
			typeMap := g.resolveGenericClassTypeArgs(gcls, e, g.currentClassName)
			mangledClass := gcName + "_" + g.mangledSuffix(typeMap, gcls.TypeParams)
			if argStr != "" {
				return fmt.Sprintf("%s_new(%s)", mangledClass, argStr)
			}
			return fmt.Sprintf("%s_new()", mangledClass)
		}
	}

	// Check if callee is a closure variable
	if ident, ok := e.Callee.(*parser.Ident); ok {
		if g.localVars != nil {
			if t, ok := g.localVars[ident.Name]; ok && t == "KlClosure*" {
				callee := g.resolveIdent(ident.Name)
				// Resolve fn type from localVarTypes for proper casting
				var fnType *parser.FnType
				if g.localVarTypes != nil {
					if te, ok := g.localVarTypes[ident.Name]; ok {
						fnType, _ = te.(*parser.FnType)
					}
				}
				return g.emitClosureCall(callee, fnType, args)
			}
		}
	}

	// Check if callee is a bare identifier that's a method of current class
	if ident, ok := e.Callee.(*parser.Ident); ok {
		if g.isMethod(ident.Name) {
			// In DLL mode inside main(), call through function pointer for hot reload
			if g.DLLMode && g.dllInsideMain && ident.Name != "main" {
				if argStr != "" {
					return fmt.Sprintf("self->_fn_%s(self, %s)", ident.Name, argStr)
				}
				return fmt.Sprintf("self->_fn_%s(self)", ident.Name)
			}
			if argStr != "" {
				return fmt.Sprintf("%s_%s(self, %s)", g.currentClassName, ident.Name, argStr)
			}
			return fmt.Sprintf("%s_%s(self)", g.currentClassName, ident.Name)
		}
		// Constructor calls for value types: vec2/vec3/vec4/quat/Color → compound literal
		switch ident.Name {
		case "vec2":
			return fmt.Sprintf("(vec2){%s}", argStr)
		case "vec3":
			return fmt.Sprintf("(vec3){%s}", argStr)
		case "vec4":
			return fmt.Sprintf("(vec4){%s}", argStr)
		case "quat":
			return fmt.Sprintf("(quat){%s}", argStr)
		case "Color":
			return fmt.Sprintf("(Color){%s}", argStr)
		case "Rectangle":
			return fmt.Sprintf("(Rectangle){%s}", argStr)
		case "Random":
			return fmt.Sprintf("kl_random_new(%s)", argStr)
		}
		// Check if it's a class constructor: Ball() → Main_Ball_new()
		fullName := g.resolveFullClassName(ident.Name, g.currentClassName)
		if _, ok := g.classes[fullName]; ok {
			if argStr != "" {
				return fmt.Sprintf("%s_new(%s)", fullName, argStr)
			}
			return fmt.Sprintf("%s_new()", fullName)
		}
		// Check active "with" modules for bare function calls
		for i := len(g.withModules) - 1; i >= 0; i-- {
			mod := g.withModules[i]
			if funcs, ok := stdlibModuleFuncs[mod]; ok {
				if cFunc, ok := funcs[ident.Name]; ok {
					if argStr != "" {
						return fmt.Sprintf("%s(%s)", cFunc, argStr)
					}
					return fmt.Sprintf("%s()", cFunc)
				}
			}
		}
	}

	// Method call on member: obj.method(args)
	if member, ok := e.Callee.(*parser.MemberExpr); ok {
		// Check if this is a stdlib module function call: math.sin(x) → sinf(x)
		if ident, ok := member.Object.(*parser.Ident); ok {
			if funcs, ok := stdlibModuleFuncs[ident.Name]; ok {
				if cFunc, ok := funcs[member.Field]; ok {
					if argStr != "" {
						return fmt.Sprintf("%s(%s)", cFunc, argStr)
					}
					return fmt.Sprintf("%s()", cFunc)
				}
			}
		}

		obj := g.exprToC(member.Object)
		methodName := member.Field

		// Built-in List methods
		if g.isListExpr(member.Object) {
			elemType := g.resolveForElemType(member.Object)
			switch methodName {
			case "append":
				return fmt.Sprintf("kl_list_push(%s, %s)", obj, g.listPushCast(argStr, elemType))
			case "count":
				return fmt.Sprintf("%s->count", obj)
			case "get":
				return g.listGetCast(elemType, obj, argStr)
			case "reverse":
				return fmt.Sprintf("kl_list_reverse(%s)", obj)
			case "clear":
				return fmt.Sprintf("kl_list_clear(%s)", obj)
			case "clone":
				return fmt.Sprintf("kl_list_clone(%s)", obj)
			case "pop":
				return g.listGetCastDirect(elemType, fmt.Sprintf("kl_list_pop(%s)", obj))
			case "first":
				return g.listGetCastDirect(elemType, fmt.Sprintf("kl_list_first(%s)", obj))
			case "last":
				return g.listGetCastDirect(elemType, fmt.Sprintf("kl_list_last(%s)", obj))
			case "remove":
				return fmt.Sprintf("kl_list_remove(%s, %s)", obj, argStr)
			case "insert":
				return fmt.Sprintf("kl_list_insert(%s, %s, %s)", obj, args[0], g.listPushCast(args[1], elemType))
			case "slice":
				return fmt.Sprintf("kl_list_slice(%s, %s)", obj, argStr)
			case "contains":
				return fmt.Sprintf("kl_list_contains(%s, %s)", obj, g.listPushCast(argStr, elemType))
			case "index_of":
				return fmt.Sprintf("kl_list_index_of(%s, %s)", obj, g.listPushCast(argStr, elemType))
			}
		}

		// Built-in KlRandom methods: rnd.rangei(0,10) → kl_random_rangei(&rnd, 0, 10)
		objType := g.resolveExprType(member.Object)
		if objType == "KlRandom" {
			cFunc := "kl_random_" + methodName
			if argStr != "" {
				return fmt.Sprintf("%s(&%s, %s)", cFunc, obj, argStr)
			}
			return fmt.Sprintf("%s(&%s)", cFunc, obj)
		}

		// Resolve the type of the object for proper method dispatch
		typeName := g.resolveExprTypeName(member.Object)
		if typeName != "" {
			castObj := fmt.Sprintf("(%s*)%s", typeName, obj)
			if argStr != "" {
				return fmt.Sprintf("%s_%s(%s, %s)", typeName, methodName, castObj, argStr)
			}
			return fmt.Sprintf("%s_%s(%s)", typeName, methodName, castObj)
		}

		if argStr != "" {
			return fmt.Sprintf("%s_%s(%s, %s)", obj, methodName, obj, argStr)
		}
		return fmt.Sprintf("%s_%s(%s)", obj, methodName, obj)
	}

	callee := g.exprToC(e.Callee)
	if argStr != "" {
		return fmt.Sprintf("%s(%s)", callee, argStr)
	}
	return fmt.Sprintf("%s()", callee)
}

func (g *Generator) emitStructLit(e *parser.StructLit) string {
	return g.emitTypedStructLit(e, "struct _anon")
}

func (g *Generator) emitTypedStructLit(e *parser.StructLit, cType string) string {
	// Pointer types: heap-allocate with RC
	if strings.HasSuffix(cType, "*") {
		structType := strings.TrimSuffix(cType, "*")
		tmpVar := fmt.Sprintf("_sl_%d", g.structLitCounter)
		g.structLitCounter++

		g.writeln("%s* %s = (%s*)kl_alloc_rc(sizeof(%s), %s_destroy);", structType, tmpVar, structType, structType, structType)

		// Set type ID and tracer
		if cls, ok := g.classes[structType]; ok && cls.Parent != "" {
			g.writeln("%s->_base._header.type_id = KLTYPE_%s;", tmpVar, structType)
			g.writeln("%s->_base._header.tracer = %s_trace;", tmpVar, structType)
		} else {
			g.writeln("%s->_header.type_id = KLTYPE_%s;", tmpVar, structType)
			g.writeln("%s->_header.tracer = %s_trace;", tmpVar, structType)
		}

		// Set fields
		for _, f := range e.Fields {
			val := g.exprToC(f.Value)
			if f.Name != "" {
				// Determine if it's a parent field
				fieldTarget := fmt.Sprintf("%s->%s", tmpVar, f.Name)
				if cls, ok := g.classes[structType]; ok && cls.Parent != "" {
					prefix := g.extractPrefix(structType, cls.Name)
					parentFullName := g.resolveParentName(prefix, cls.Parent)
					if g.isFieldInAncestor(parentFullName, f.Name) {
						fieldTarget = fmt.Sprintf("%s->_base.%s", tmpVar, f.Name)
					}
				}
				g.writeln("%s = %s;", fieldTarget, val)
			}
		}

		return tmpVar
	}

	// Non-pointer compound literal
	fields := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		if f.Name != "" {
			fields[i] = fmt.Sprintf(".%s = %s", f.Name, g.exprToC(f.Value))
		} else {
			fields[i] = g.exprToC(f.Value)
		}
	}
	return fmt.Sprintf("(%s){%s}", cType, strings.Join(fields, ", "))
}

func (g *Generator) extractPrefix(fullName, shortName string) string {
	suffix := "_" + shortName
	if strings.HasSuffix(fullName, suffix) {
		return fullName[:len(fullName)-len(suffix)]
	}
	return ""
}

func (g *Generator) isFieldInAncestor(className, fieldName string) bool {
	if cls, ok := g.classes[className]; ok {
		for _, f := range cls.Fields {
			if f.Name == fieldName {
				return true
			}
		}
		if cls.Parent != "" {
			prefix := g.extractPrefix(className, cls.Name)
			parentName := g.resolveParentName(prefix, cls.Parent)
			return g.isFieldInAncestor(parentName, fieldName)
		}
	}
	return false
}

func (g *Generator) emitArrayLit(e *parser.ArrayLit) string {
	// Without type context, default to non-RC items
	return "kl_list_new(false)"
}

// --- Lambda/closure support ---

func (g *Generator) emitLambda(e *parser.LambdaExpr) string {
	id := g.lambdaCounter
	g.lambdaCounter++
	lambdaName := fmt.Sprintf("_lambda_%d", id)
	capsName := fmt.Sprintf("_lambda_%d_captures", id)

	// Determine captured variables: scan lambda body for identifiers
	// that reference local vars in the enclosing scope
	captures := g.findCaptures(e.Body)

	// Build the lambda return type from params (for now, void if no return statements found)
	retType := "void"

	// Emit capture struct into lambdaDefs buffer
	if len(captures) > 0 {
		fmt.Fprintf(&g.lambdaDefs, "typedef struct {\n")
		for _, cap := range captures {
			cType := "int" // default
			if g.localVars != nil {
				if t, ok := g.localVars[cap]; ok && t != "" {
					cType = t
				}
			}
			fmt.Fprintf(&g.lambdaDefs, "    %s %s;\n", cType, cap)
		}
		fmt.Fprintf(&g.lambdaDefs, "} %s;\n\n", capsName)
	}

	// Emit lambda function into lambdaDefs buffer
	fmt.Fprintf(&g.lambdaDefs, "%s %s(void* _cap", retType, lambdaName)
	for _, p := range e.Params {
		pType := "int"
		if p.TypeExpr != nil {
			pType = g.typeToC(p.TypeExpr, g.currentClassName)
		}
		fmt.Fprintf(&g.lambdaDefs, ", %s %s", pType, p.Name)
	}
	fmt.Fprintf(&g.lambdaDefs, ") {\n")
	if len(captures) > 0 {
		fmt.Fprintf(&g.lambdaDefs, "    %s* captures = (%s*)_cap;\n", capsName, capsName)
	}

	// Emit lambda body — we need to redirect identifiers to captures->field
	// Save and set up a new local vars context for the lambda body
	prevLocalVars := g.localVars
	prevLocalVarTypes := g.localVarTypes
	g.localVars = map[string]string{}
	g.localVarTypes = map[string]parser.TypeExpr{}

	// Register lambda params as locals
	for _, p := range e.Params {
		pType := "int"
		if p.TypeExpr != nil {
			pType = g.typeToC(p.TypeExpr, g.currentClassName)
		}
		g.localVars[p.Name] = pType
	}

	// Generate body into a temporary buffer, then write to lambdaDefs
	prevOut := g.out
	prevIndent := g.indent
	g.out = strings.Builder{}
	g.indent = 1

	// For captured vars, we need special handling: register them as locals
	// pointing to "captures->varname" so resolveIdent picks them up
	capturedVarMap := map[string]string{}
	for _, cap := range captures {
		cType := "int"
		if prevLocalVars != nil {
			if t, ok := prevLocalVars[cap]; ok && t != "" {
				cType = t
			}
		}
		capturedVarMap[cap] = cType
		// Add a local var declaration that aliases captures->field
		g.localVars[cap] = cType
	}

	if e.Body != nil {
		// We need to emit the body but replace captured var references
		// with captures->varname. We'll emit local aliases at the top.
		for _, cap := range captures {
			cType := capturedVarMap[cap]
			g.writeln("%s %s = captures->%s;", cType, cap, cap)
		}
		g.emitBlock(e.Body, g.currentClassName)
	}

	lambdaBody := g.out.String()
	g.out = prevOut
	g.indent = prevIndent
	g.localVars = prevLocalVars
	g.localVarTypes = prevLocalVarTypes

	fmt.Fprintf(&g.lambdaDefs, "%s", lambdaBody)
	fmt.Fprintf(&g.lambdaDefs, "}\n\n")

	// Emit inline code: allocate captures, fill them, create KlClosure
	// We return a multi-statement expression using a statement-expression GCC extension
	// Actually, we need to emit statements before this expression. Use a helper pattern:
	// Emit setup statements into the current method output, return the final variable name.
	closureVar := fmt.Sprintf("_cl_%d", id)

	if len(captures) > 0 {
		capVar := fmt.Sprintf("_cap_%d", id)
		g.writeln("%s* %s = (%s*)kl_alloc(sizeof(%s));", capsName, capVar, capsName, capsName)
		for _, cap := range captures {
			capCType := "int"
			if g.localVars != nil {
				if t, ok := g.localVars[cap]; ok && t != "" {
					capCType = t
				}
			}
			g.writeln("%s->%s = %s;", capVar, cap, g.resolveIdent(cap))
			if g.isRefCountedType(capCType) {
				g.writeln("kl_retain(%s->%s);", capVar, cap)
			}
		}
		g.writeln("KlClosure* %s = (KlClosure*)kl_alloc_rc(sizeof(KlClosure), kl_closure_destroy);", closureVar)
		g.writeln("%s->fn = (void*)%s;", closureVar, lambdaName)
		g.writeln("%s->captures = %s;", closureVar, capVar)
		g.writeln("%s->captures_dtor = NULL;", closureVar)
	} else {
		g.writeln("KlClosure* %s = (KlClosure*)kl_alloc_rc(sizeof(KlClosure), kl_closure_destroy);", closureVar)
		g.writeln("%s->fn = (void*)%s;", closureVar, lambdaName)
		g.writeln("%s->captures = NULL;", closureVar)
		g.writeln("%s->captures_dtor = NULL;", closureVar)
	}

	return closureVar
}

// findCaptures scans a block for identifiers that reference enclosing-scope local vars.
func (g *Generator) findCaptures(block *parser.Block) []string {
	if block == nil || g.localVars == nil {
		return nil
	}
	seen := map[string]bool{}
	var captures []string
	g.findCapturesInBlock(block, seen, &captures)
	return captures
}

func (g *Generator) findCapturesInBlock(block *parser.Block, seen map[string]bool, captures *[]string) {
	for _, stmt := range block.Stmts {
		g.findCapturesInStmt(stmt, seen, captures)
	}
}

func (g *Generator) findCapturesInStmt(stmt parser.Stmt, seen map[string]bool, captures *[]string) {
	switch s := stmt.(type) {
	case *parser.ExprStmt:
		g.findCapturesInExpr(s.Expr, seen, captures)
	case *parser.VarDecl:
		if s.Value != nil {
			g.findCapturesInExpr(s.Value, seen, captures)
		}
	case *parser.ReturnStmt:
		if s.Value != nil {
			g.findCapturesInExpr(s.Value, seen, captures)
		}
	case *parser.IfStmt:
		g.findCapturesInExpr(s.Condition, seen, captures)
		if s.Then != nil {
			g.findCapturesInBlock(s.Then, seen, captures)
		}
		if s.ThenStmt != nil {
			g.findCapturesInStmt(s.ThenStmt, seen, captures)
		}
		if s.Else != nil {
			g.findCapturesInStmt(s.Else, seen, captures)
		}
	case *parser.ForStmt:
		g.findCapturesInExpr(s.Iterable, seen, captures)
		if s.Body != nil {
			g.findCapturesInBlock(s.Body, seen, captures)
		}
	case *parser.WhileStmt:
		g.findCapturesInExpr(s.Condition, seen, captures)
		if s.Body != nil {
			g.findCapturesInBlock(s.Body, seen, captures)
		}
	case *parser.AssignStmt:
		g.findCapturesInExpr(s.Target, seen, captures)
		g.findCapturesInExpr(s.Value, seen, captures)
	case *parser.Block:
		g.findCapturesInBlock(s, seen, captures)
	}
}

func (g *Generator) findCapturesInExpr(expr parser.Expr, seen map[string]bool, captures *[]string) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *parser.Ident:
		// Check if this identifier references a local var in the enclosing scope
		if g.localVars != nil {
			if _, ok := g.localVars[e.Name]; ok && !seen[e.Name] {
				seen[e.Name] = true
				*captures = append(*captures, e.Name)
			}
		}
	case *parser.BinaryExpr:
		g.findCapturesInExpr(e.Left, seen, captures)
		g.findCapturesInExpr(e.Right, seen, captures)
	case *parser.UnaryExpr:
		g.findCapturesInExpr(e.Operand, seen, captures)
	case *parser.CallExpr:
		g.findCapturesInExpr(e.Callee, seen, captures)
		for _, arg := range e.Args {
			g.findCapturesInExpr(arg, seen, captures)
		}
	case *parser.MemberExpr:
		g.findCapturesInExpr(e.Object, seen, captures)
	case *parser.IndexExpr:
		g.findCapturesInExpr(e.Object, seen, captures)
		g.findCapturesInExpr(e.Index, seen, captures)
	case *parser.NullCoalesce:
		g.findCapturesInExpr(e.Left, seen, captures)
		g.findCapturesInExpr(e.Right, seen, captures)
	}
}

// emitClosureCall generates the C code for calling a KlClosure variable.
func (g *Generator) emitClosureCall(callee string, fnType *parser.FnType, args []string) string {
	// Build the function pointer cast: ((retType(*)(void*, paramTypes...))var->fn)
	retType := "void"
	if fnType != nil && fnType.ReturnType != nil {
		retType = g.typeToC(fnType.ReturnType, g.currentClassName)
	}

	paramTypes := "void*"
	if fnType != nil && len(fnType.ParamTypes) > 0 {
		pts := []string{"void*"}
		for _, pt := range fnType.ParamTypes {
			pts = append(pts, g.typeToC(pt, g.currentClassName))
		}
		paramTypes = strings.Join(pts, ", ")
	}

	callArgs := callee + "->captures"
	if len(args) > 0 {
		callArgs += ", " + strings.Join(args, ", ")
	}

	return fmt.Sprintf("((%s(*)(%s))%s->fn)(%s)", retType, paramTypes, callee, callArgs)
}

// --- Type helpers ---

func (g *Generator) typeToC(t parser.TypeExpr, context string) string {
	if t == nil {
		return "void"
	}
	switch ty := t.(type) {
	case *parser.SimpleType:
		// Check generic type substitutions first
		if g.typeSubstitutions != nil {
			if concrete, ok := g.typeSubstitutions[ty.Name]; ok {
				return concrete
			}
		}
		return g.simpleTypeToCWithContext(ty.Name, context)
	case *parser.GenericType:
		if ty.Name == "List" {
			return "KlList*"
		}
		return ty.Name + "*"
	case *parser.UnionType:
		return "void*" // placeholder — tagged unions come later
	case *parser.InlineClassType:
		return "void*"
	case *parser.FnType:
		return "KlClosure*"
	}
	return "void"
}

func (g *Generator) simpleTypeToCWithContext(name, context string) string {
	switch name {
	case "int":
		return "int"
	case "float":
		return "float"
	case "bool":
		return "bool"
	case "string":
		return "const char*"
	case "vec2":
		return "vec2"
	case "vec3":
		return "vec3"
	case "vec4":
		return "vec4"
	case "mat4":
		return "mat4"
	case "quat":
		return "quat"
	case "Color":
		return "Color"
	case "Rectangle":
		return "Rectangle"
	case "Texture2D", "Texture":
		return "Texture2D"
	case "Font":
		return "Font"
	case "Sound":
		return "Sound"
	case "Music":
		return "Music"
	case "Camera2D":
		return "Camera2D"
	case "Camera3D":
		return "Camera3D"
	case "Random":
		return "KlRandom"
	default:
		// Try to resolve as a nested class
		if context != "" {
			full := context + "_" + name
			if _, ok := g.classes[full]; ok {
				return full + "*"
			}
		}
		// Try as sibling class (strip last segment of context, try parent_name)
		if context != "" {
			if idx := strings.LastIndex(context, "_"); idx >= 0 {
				siblingFull := context[:idx] + "_" + name
				if _, ok := g.classes[siblingFull]; ok {
					return siblingFull + "*"
				}
			}
		}
		// Try as top-level
		if _, ok := g.classes[name]; ok {
			return name + "*"
		}
		return name + "*"
	}
}

// --- Memory management helpers ---

func (g *Generator) listItemsAreRC(typeExpr parser.TypeExpr, className string) bool {
	if typeExpr == nil {
		return false
	}
	if gt, ok := typeExpr.(*parser.GenericType); ok && gt.Name == "List" && len(gt.TypeArgs) > 0 {
		elemCType := g.typeToC(gt.TypeArgs[0], className)
		return g.isRefCountedType(elemCType)
	}
	return false
}

func (g *Generator) boolToC(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (g *Generator) resolveAssignTargetType(target parser.Expr, className string) string {
	switch t := target.(type) {
	case *parser.Ident:
		// Check local vars
		if g.localVars != nil {
			if cType, ok := g.localVars[t.Name]; ok {
				return cType
			}
		}
		// Check class fields
		if g.currentClass != nil {
			for _, f := range g.currentClass.Fields {
				if f.Name == t.Name {
					return g.fieldCType(f, g.currentClassName)
				}
			}
		}
	case *parser.MemberExpr:
		// Resolve the type of the object, then look up the field
		typeName := g.resolveExprTypeName(t.Object)
		if typeName != "" {
			if cls, ok := g.classes[typeName]; ok {
				for _, f := range cls.Fields {
					if f.Name == t.Field {
						return g.fieldCType(f, typeName)
					}
				}
				// Check parent
				if cls.Parent != "" {
					parentName := g.resolveParentName("", cls.Parent)
					if parent, ok := g.classes[parentName]; ok {
						for _, f := range parent.Fields {
							if f.Name == t.Field {
								return g.fieldCType(f, parentName)
							}
						}
					}
				}
			}
		}
	case *parser.IndexExpr:
		// list[i] = val → element type
		return g.resolveForElemType(t.Object)
	}
	return ""
}

func (g *Generator) isWeakAssignTarget(target parser.Expr, className string) bool {
	switch t := target.(type) {
	case *parser.Ident:
		// Bare identifier — could be a field on current class
		if g.localVars != nil {
			if _, ok := g.localVars[t.Name]; ok {
				return false // local vars are never weak
			}
		}
		if g.currentClass != nil {
			return g.isWeakField(g.currentClassName, t.Name)
		}
	case *parser.MemberExpr:
		typeName := g.resolveExprTypeName(t.Object)
		if typeName != "" {
			return g.isWeakField(typeName, t.Field)
		}
	}
	return false
}

func (g *Generator) identifyReturnedLocalVar(expr parser.Expr) string {
	if ident, ok := expr.(*parser.Ident); ok {
		if g.localVars != nil {
			if _, ok := g.localVars[ident.Name]; ok {
				return ident.Name
			}
		}
	}
	return ""
}

func (g *Generator) returnTypeToC(t parser.TypeExpr, context string) string {
	if t == nil {
		return "void"
	}
	return g.typeToC(t, context)
}

func (g *Generator) inferCType(expr parser.Expr) string {
	if expr == nil {
		return "int"
	}
	switch e := expr.(type) {
	case *parser.IntLit:
		return "int"
	case *parser.FloatLit:
		return "float"
	case *parser.StringLit:
		return "const char*"
	case *parser.BoolLit:
		return "bool"
	case *parser.UnaryExpr:
		return g.inferCType(e.Operand)
	case *parser.CallExpr:
		// Check if it's a value type constructor
		if ident, ok := e.Callee.(*parser.Ident); ok {
			switch ident.Name {
			case "vec2":
				return "vec2"
			case "vec3":
				return "vec3"
			case "vec4":
				return "vec4"
			case "quat":
				return "quat"
			case "Color":
				return "Color"
			case "Rectangle":
				return "Rectangle"
			case "Random":
				return "KlRandom"
			}
			// Check if it's a known vector/math global function
			if rt := g.inferVectorFuncReturnType(ident.Name); rt != "" {
				return rt
			}
			// Check active "with" modules for bare function call return types
			for i := len(g.withModules) - 1; i >= 0; i-- {
				mod := g.withModules[i]
				if funcs, ok := stdlibModuleFuncs[mod]; ok {
					if _, ok := funcs[ident.Name]; ok {
						if mod == "rl" {
							return g.inferRlReturnType(ident.Name)
						}
						if mod == "math" {
							return "float"
						}
					}
				}
			}
			// Check if it's a class constructor → return pointer type
			fullName := g.resolveFullClassName(ident.Name, g.currentClassName)
			if _, ok := g.classes[fullName]; ok {
				return fullName + "*"
			}
			// Check if it's a generic class constructor → return the mangled class pointer type
			gcName := g.resolveGenericClassName(ident.Name, g.currentClassName)
			if gcls, ok := g.genericClasses[gcName]; ok {
				typeMap := g.resolveGenericClassTypeArgs(gcls, e, g.currentClassName)
				mangledClass := gcName + "_" + g.mangledSuffix(typeMap, gcls.TypeParams)
				return mangledClass + "*"
			}
		}
		// Check if it's a module function call — infer return type
		if member, ok := e.Callee.(*parser.MemberExpr); ok {
			if ident, ok := member.Object.(*parser.Ident); ok {
				if funcs, ok := stdlibModuleFuncs[ident.Name]; ok {
					if _, ok := funcs[member.Field]; ok {
						if ident.Name == "math" {
							return "float"
						}
						if ident.Name == "io" {
							switch member.Field {
							case "read_file":
								return "const char*"
							case "file_exists", "dir_exists", "write_file", "append_file", "delete_file", "create_dir":
								return "bool"
							case "list_dir":
								return "KlList*"
							}
						}
						if ident.Name == "rl" {
							return g.inferRlReturnType(member.Field)
						}
					}
				}
			}
			// Check if it's a list method call — infer return type
			if g.isListExpr(member.Object) {
				switch member.Field {
				case "clone", "filter", "slice", "map":
					return "KlList*"
				case "first", "last", "pop", "find":
					return g.resolveForElemType(member.Object)
				case "count", "find_index", "index_of":
					return "int"
				case "contains":
					return "bool"
				}
			}
		}
		return "int"
	case *parser.IndexExpr:
		return g.resolveForElemType(e.Object)
	case *parser.LambdaExpr:
		return "KlClosure*"
	}
	return "int"
}

func (g *Generator) constructorParams(cls *parser.ClassDecl, className string) string {
	if cls.Constructor == nil {
		return "void"
	}
	params := make([]string, len(cls.Constructor.Params))
	for i, p := range cls.Constructor.Params {
		params[i] = fmt.Sprintf("%s %s", g.typeToC(p.TypeExpr, className), p.Name)
	}
	if len(params) == 0 {
		return "void"
	}
	return strings.Join(params, ", ")
}

func (g *Generator) methodParams(className string, method *parser.MethodDecl) string {
	parts := []string{fmt.Sprintf("%s* self", className)}
	for _, p := range method.Params {
		parts = append(parts, fmt.Sprintf("%s %s", g.typeToC(p.TypeExpr, className), p.Name))
	}
	return strings.Join(parts, ", ")
}

func (g *Generator) resolveParentName(prefix, parent string) string {
	if prefix != "" {
		full := prefix + "_" + parent
		if _, ok := g.classes[full]; ok {
			return full
		}
	}
	return parent
}

// --- Type ID helpers ---

func (g *Generator) emitTypeIds() {
	g.writeln("enum {")
	g.indent++
	id := 1
	for _, cls := range g.allClasses() {
		g.emitTypeIdEnum("", cls, &id)
	}
	g.indent--
	g.writeln("};")
	g.writeln("")
}

func (g *Generator) emitTypeIdEnum(prefix string, cls *parser.ClassDecl, id *int) {
	if len(cls.TypeParams) > 0 {
		return // generic template — monomorphized versions get their own type IDs
	}
	name := g.cName(prefix, cls.Name)
	g.writeln("KLTYPE_%s = %d,", name, *id)
	*id++
	for _, nested := range cls.Classes {
		g.emitTypeIdEnum(name, nested, id)
	}
}

func (g *Generator) resolveFullClassName(shortName, context string) string {
	// Direct match
	if _, ok := g.classes[shortName]; ok {
		return shortName
	}
	// Try with context prefix
	if context != "" {
		full := context + "_" + shortName
		if _, ok := g.classes[full]; ok {
			return full
		}
	}
	// Search all classes for a suffix match
	for fullName := range g.classes {
		if strings.HasSuffix(fullName, "_"+shortName) {
			return fullName
		}
	}
	return shortName
}

// --- Output helpers ---

func (g *Generator) writeln(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	for i := 0; i < g.indent; i++ {
		g.out.WriteString("    ")
	}
	g.out.WriteString(line)
	g.out.WriteString("\n")
}

func (g *Generator) write(format string, args ...interface{}) {
	line := fmt.Sprintf(format, args...)
	for i := 0; i < g.indent; i++ {
		g.out.WriteString("    ")
	}
	g.out.WriteString(line)
}

// --- Exported methods for LSP analysis ---

// GetClasses returns the registered class map (className → *ClassDecl).
func (g *Generator) GetClasses() map[string]*parser.ClassDecl {
	return g.classes
}

// FieldCType resolves the C type of a field declaration.
func (g *Generator) FieldCType(field *parser.FieldDecl, className string) string {
	return g.fieldCType(field, className)
}

// TypeToC converts a Klang type expression to its C representation.
func (g *Generator) TypeToC(t parser.TypeExpr, context string) string {
	return g.typeToC(t, context)
}

// InferCType infers the C type of an expression (safe for LSP — recovers from panics).
func (g *Generator) InferCType(expr parser.Expr) (result string) {
	defer func() {
		if r := recover(); r != nil {
			result = ""
		}
	}()
	return g.inferCType(expr)
}

// FindFieldInClass checks if a class (or its parents) has a field with the given name.
func (g *Generator) FindFieldInClass(cls *parser.ClassDecl, name string) bool {
	return g.findFieldInClass(cls, name)
}

// FindMethodInClass checks if a class (or its parents) has a method with the given name.
func (g *Generator) FindMethodInClass(cls *parser.ClassDecl, name string) bool {
	return g.findMethodInClass(cls, name)
}

// IsValueType returns true if the C type is a value type (stack-allocated, uses . not ->).
func (g *Generator) IsValueType(cType string) bool {
	return g.isValueType(cType)
}

// StdlibModuleFuncs returns the stdlib module→function→C mapping.
func StdlibModuleFuncs() map[string]map[string]string {
	return stdlibModuleFuncs
}

// StdlibModuleConstants returns the stdlib module→constant→C mapping.
func StdlibModuleConstants() map[string]map[string]string {
	return stdlibModuleConstants
}

// StdlibConstNamespaces returns the namespace→constant→C mapping (Colors, Key, etc.).
func StdlibConstNamespaces() map[string]map[string]string {
	return stdlibConstNamespaces
}
