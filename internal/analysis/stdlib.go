package analysis

// StdlibFunc holds signature info for a stdlib function.
type StdlibFunc struct {
	Name   string
	Detail string // e.g. "sin(x:float):float"
}

// StdlibModuleSignatures maps module name → function signatures for IntelliSense.
// Core modules are defined here; vendor libs contribute via RegisterVendor.
var StdlibModuleSignatures = map[string][]StdlibFunc{
	"math": {
		{"sin", "sin(x:float):float"},
		{"cos", "cos(x:float):float"},
		{"tan", "tan(x:float):float"},
		{"asin", "asin(x:float):float"},
		{"acos", "acos(x:float):float"},
		{"atan", "atan(x:float):float"},
		{"atan2", "atan2(y:float, x:float):float"},
		{"sqrt", "sqrt(x:float):float"},
		{"pow", "pow(base:float, exp:float):float"},
		{"abs", "abs(x:float):float"},
		{"floor", "floor(x:float):float"},
		{"ceil", "ceil(x:float):float"},
		{"round", "round(x:float):float"},
		{"min", "min(a:float, b:float):float"},
		{"max", "max(a:float, b:float):float"},
		{"clamp", "clamp(value:float, min:float, max:float):float"},
		{"lerp", "lerp(a:float, b:float, t:float):float"},
		{"sign", "sign(x:float):float"},
		{"deg2rad", "deg2rad(degrees:float):float"},
		{"rad2deg", "rad2deg(radians:float):float"},
	},
	"io": {
		{"read_file", "read_file(path:string):string"},
		{"write_file", "write_file(path:string, content:string):bool"},
		{"append_file", "append_file(path:string, content:string):bool"},
		{"file_exists", "file_exists(path:string):bool"},
		{"delete_file", "delete_file(path:string):bool"},
		{"create_dir", "create_dir(path:string):bool"},
		{"dir_exists", "dir_exists(path:string):bool"},
		{"list_dir", "list_dir(path:string):List"},
	},
}

// StdlibModuleConstantNames maps module name → constant names.
var StdlibModuleConstantNames = map[string][]StdlibFunc{
	"math": {
		{"PI", "float — 3.14159..."},
		{"TAU", "float — 6.28318..."},
		{"E", "float — 2.71828..."},
		{"DEG2RAD", "float"},
		{"RAD2DEG", "float"},
		{"INF", "float — infinity"},
		{"EPSILON", "float"},
	},
}

// StdlibNamespaces maps namespace → members (core only; vendors add more).
var StdlibNamespaces = map[string][]StdlibFunc{}

// Keywords for completion.
var Keywords = []string{
	"if", "else", "for", "while", "return", "with",
	"class", "enum", "event", "new", "fn",
	"true", "false", "this", "not", "and", "or", "is", "in",
}

// BuiltinTypes for type completion (core types only; vendors add more).
var BuiltinTypes = []string{
	"int", "float", "bool", "string",
	"vec2", "vec3", "vec4", "mat4", "quat",
	"List", "Dictionary", "fn",
}

// BuiltinTypeMembers maps built-in value types to their fields and methods.
var BuiltinTypeMembers = map[string][]CompletionItem{
	"vec2": {
		{Label: "x", Detail: "float", Kind: CompletionKindField},
		{Label: "y", Detail: "float", Kind: CompletionKindField},
	},
	"vec3": {
		{Label: "x", Detail: "float", Kind: CompletionKindField},
		{Label: "y", Detail: "float", Kind: CompletionKindField},
		{Label: "z", Detail: "float", Kind: CompletionKindField},
	},
	"vec4": {
		{Label: "x", Detail: "float", Kind: CompletionKindField},
		{Label: "y", Detail: "float", Kind: CompletionKindField},
		{Label: "z", Detail: "float", Kind: CompletionKindField},
		{Label: "w", Detail: "float", Kind: CompletionKindField},
	},
	"quat": {
		{Label: "x", Detail: "float", Kind: CompletionKindField},
		{Label: "y", Detail: "float", Kind: CompletionKindField},
		{Label: "z", Detail: "float", Kind: CompletionKindField},
		{Label: "w", Detail: "float", Kind: CompletionKindField},
	},
	"List": {
		{Label: "append", Detail: "append(item)", Kind: CompletionKindMethod},
		{Label: "count", Detail: "count():int", Kind: CompletionKindMethod},
		{Label: "get", Detail: "get(index:int):T", Kind: CompletionKindMethod},
		{Label: "remove", Detail: "remove(index:int)", Kind: CompletionKindMethod},
		{Label: "remove_all", Detail: "remove_all(pred:fn(T):bool)", Kind: CompletionKindMethod},
		{Label: "insert", Detail: "insert(index:int, item)", Kind: CompletionKindMethod},
		{Label: "pop", Detail: "pop():T", Kind: CompletionKindMethod},
		{Label: "first", Detail: "first():T", Kind: CompletionKindMethod},
		{Label: "last", Detail: "last():T", Kind: CompletionKindMethod},
		{Label: "clear", Detail: "clear()", Kind: CompletionKindMethod},
		{Label: "reverse", Detail: "reverse()", Kind: CompletionKindMethod},
		{Label: "clone", Detail: "clone():List<T>", Kind: CompletionKindMethod},
		{Label: "slice", Detail: "slice(start:int, end:int):List<T>", Kind: CompletionKindMethod},
		{Label: "contains", Detail: "contains(item):bool", Kind: CompletionKindMethod},
		{Label: "index_of", Detail: "index_of(item):int", Kind: CompletionKindMethod},
		{Label: "sort", Detail: "sort(cmp:fn(a,b):float)", Kind: CompletionKindMethod},
		{Label: "sort_by", Detail: "sort_by(key:fn(T):float)", Kind: CompletionKindMethod},
		{Label: "filter", Detail: "filter(pred:fn(T):bool):List<T>", Kind: CompletionKindMethod},
		{Label: "map", Detail: "map(fn:fn(T):U):List<U>", Kind: CompletionKindMethod},
		{Label: "find", Detail: "find(pred:fn(T):bool):T", Kind: CompletionKindMethod},
		{Label: "find_index", Detail: "find_index(pred:fn(T):bool):int", Kind: CompletionKindMethod},
	},
	"Dictionary": {
		{Label: "append", Detail: "append(key:K, value:V)", Kind: CompletionKindMethod},
		{Label: "set", Detail: "set(key:K, value:V)", Kind: CompletionKindMethod},
		{Label: "get", Detail: "get(key:K):V", Kind: CompletionKindMethod},
		{Label: "has", Detail: "has(key:K):bool", Kind: CompletionKindMethod},
		{Label: "remove", Detail: "remove(key:K)", Kind: CompletionKindMethod},
		{Label: "keys", Detail: "keys():List<K>", Kind: CompletionKindMethod},
		{Label: "values", Detail: "values():List<V>", Kind: CompletionKindMethod},
		{Label: "count", Detail: "count():int", Kind: CompletionKindMethod},
		{Label: "clear", Detail: "clear()", Kind: CompletionKindMethod},
	},
}

// BuiltinTypeFieldTypes maps built-in type fields to their types (for chained resolution).
var BuiltinTypeFieldTypes = map[string]map[string]string{
	"vec2": {"x": "float", "y": "float"},
	"vec3": {"x": "float", "y": "float", "z": "float"},
	"vec4": {"x": "float", "y": "float", "z": "float", "w": "float"},
	"quat": {"x": "float", "y": "float", "z": "float", "w": "float"},
}

// ModuleNames for completion (core only; vendors add more via ensureVendorsMerged).
var ModuleNames = []string{"math", "io"}

// NamespaceNames for completion (core has none; vendors add via ensureVendorsMerged).
var NamespaceNames []string

// ModuleNamespaceMap maps module name → namespace names accessible via that module.
// e.g. "rl" → ["CameraMode", "Flag", ...] so rl.CameraMode. completions work.
var ModuleNamespaceMap = map[string][]string{}
