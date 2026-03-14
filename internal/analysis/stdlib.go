package analysis

// StdlibFunc holds signature info for a stdlib function.
type StdlibFunc struct {
	Name   string
	Detail string // e.g. "sin(x:float):float"
}

// StdlibModuleSignatures maps module name → function signatures for IntelliSense.
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
	"rl": {
		// Window
		{"init_window", "init_window(width:int, height:int, title:string)"},
		{"close_window", "close_window()"},
		{"window_should_close", "window_should_close():bool"},
		{"set_target_fps", "set_target_fps(fps:int)"},
		{"get_screen_width", "get_screen_width():int"},
		{"get_screen_height", "get_screen_height():int"},
		{"get_screen_size", "get_screen_size():vec2"},
		{"toggle_fullscreen", "toggle_fullscreen()"},
		{"is_window_resized", "is_window_resized():bool"},
		{"set_window_title", "set_window_title(title:string)"},
		{"set_window_size", "set_window_size(width:int, height:int)"},
		{"get_frame_time", "get_frame_time():float"},
		{"get_time", "get_time():float"},
		{"get_fps", "get_fps():int"},
		// Drawing
		{"begin_drawing", "begin_drawing()"},
		{"end_drawing", "end_drawing()"},
		{"clear_background", "clear_background(color:Color)"},
		{"begin_mode_2d", "begin_mode_2d(camera:Camera2D)"},
		{"end_mode_2d", "end_mode_2d()"},
		{"begin_mode_3d", "begin_mode_3d(camera:Camera3D)"},
		{"end_mode_3d", "end_mode_3d()"},
		// Shapes
		{"draw_line", "draw_line(x1:int, y1:int, x2:int, y2:int, color:Color)"},
		{"draw_line_v", "draw_line_v(start:vec2, end:vec2, color:Color)"},
		{"draw_circle", "draw_circle(x:int, y:int, radius:float, color:Color)"},
		{"draw_circle_v", "draw_circle_v(center:vec2, radius:float, color:Color)"},
		{"draw_rectangle", "draw_rectangle(x:int, y:int, w:int, h:int, color:Color)"},
		{"draw_rectangle_v", "draw_rectangle_v(pos:vec2, size:vec2, color:Color)"},
		{"draw_rectangle_rec", "draw_rectangle_rec(rec:Rectangle, color:Color)"},
		{"draw_rectangle_lines", "draw_rectangle_lines(x:int, y:int, w:int, h:int, color:Color)"},
		{"draw_triangle", "draw_triangle(v1:vec2, v2:vec2, v3:vec2, color:Color)"},
		// Text
		{"draw_text", "draw_text(text:string, x:int, y:int, size:int, color:Color)"},
		{"draw_text_ex", "draw_text_ex(font:Font, text:string, pos:vec2, size:float, spacing:float, color:Color)"},
		{"measure_text", "measure_text(text:string, size:int):int"},
		{"load_font", "load_font(path:string):Font"},
		{"unload_font", "unload_font(font:Font)"},
		// Textures
		{"load_texture", "load_texture(path:string):Texture2D"},
		{"unload_texture", "unload_texture(texture:Texture2D)"},
		{"draw_texture", "draw_texture(texture:Texture2D, x:int, y:int, tint:Color)"},
		{"draw_texture_v", "draw_texture_v(texture:Texture2D, pos:vec2, tint:Color)"},
		{"draw_texture_ex", "draw_texture_ex(texture:Texture2D, pos:vec2, rotation:float, scale:float, tint:Color)"},
		{"draw_texture_rec", "draw_texture_rec(texture:Texture2D, source:Rectangle, pos:vec2, tint:Color)"},
		{"draw_texture_pro", "draw_texture_pro(texture:Texture2D, source:Rectangle, dest:Rectangle, origin:vec2, rotation:float, tint:Color)"},
		// Input — keyboard
		{"is_key_pressed", "is_key_pressed(key:int):bool"},
		{"is_key_down", "is_key_down(key:int):bool"},
		{"is_key_released", "is_key_released(key:int):bool"},
		{"is_key_up", "is_key_up(key:int):bool"},
		// Input — mouse
		{"is_mouse_button_pressed", "is_mouse_button_pressed(button:int):bool"},
		{"is_mouse_button_down", "is_mouse_button_down(button:int):bool"},
		{"is_mouse_button_released", "is_mouse_button_released(button:int):bool"},
		{"get_mouse_position", "get_mouse_position():vec2"},
		{"get_mouse_x", "get_mouse_x():int"},
		{"get_mouse_y", "get_mouse_y():int"},
		{"get_mouse_wheel_move", "get_mouse_wheel_move():float"},
		// Input — gamepad
		{"is_gamepad_available", "is_gamepad_available(gamepad:int):bool"},
		{"is_gamepad_button_pressed", "is_gamepad_button_pressed(gamepad:int, button:int):bool"},
		{"is_gamepad_button_down", "is_gamepad_button_down(gamepad:int, button:int):bool"},
		{"get_gamepad_axis_movement", "get_gamepad_axis_movement(gamepad:int, axis:int):float"},
		// Audio
		{"init_audio_device", "init_audio_device()"},
		{"close_audio_device", "close_audio_device()"},
		{"load_sound", "load_sound(path:string):Sound"},
		{"play_sound", "play_sound(sound:Sound)"},
		{"load_music_stream", "load_music_stream(path:string):Sound"},
		{"play_music_stream", "play_music_stream(music:Sound)"},
		{"update_music_stream", "update_music_stream(music:Sound)"},
		{"stop_music_stream", "stop_music_stream(music:Sound)"},
		// Helpers
		{"color", "color(r:int, g:int, b:int, a:int):Color"},
		{"color_rgb", "color_rgb(r:int, g:int, b:int):Color"},
		{"rect", "rect(x:float, y:float, w:float, h:float):Rectangle"},
		{"camera2d", "camera2d(offset:vec2, target:vec2, rotation:float, zoom:float):Camera2D"},
		{"camera3d", "camera3d(position:vec3, target:vec3, up:vec3, fovy:float):Camera3D"},
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

// StdlibNamespaces maps namespace → members (Colors, Key, Mouse, Gamepad).
var StdlibNamespaces = map[string][]StdlibFunc{
	"Colors": {
		{"LightGray", "Color"}, {"Gray", "Color"}, {"DarkGray", "Color"},
		{"Yellow", "Color"}, {"Gold", "Color"}, {"Orange", "Color"},
		{"Pink", "Color"}, {"Red", "Color"}, {"Maroon", "Color"},
		{"Green", "Color"}, {"Lime", "Color"}, {"DarkGreen", "Color"},
		{"SkyBlue", "Color"}, {"Blue", "Color"}, {"DarkBlue", "Color"},
		{"Purple", "Color"}, {"Violet", "Color"}, {"DarkPurple", "Color"},
		{"Beige", "Color"}, {"Brown", "Color"}, {"DarkBrown", "Color"},
		{"White", "Color"}, {"Black", "Color"}, {"Blank", "Color"},
		{"Magenta", "Color"}, {"RayWhite", "Color"},
	},
	"Key": {
		{"Space", "int"}, {"Enter", "int"}, {"Escape", "int"}, {"Backspace", "int"}, {"Tab", "int"},
		{"Up", "int"}, {"Down", "int"}, {"Left", "int"}, {"Right", "int"},
		{"A", "int"}, {"B", "int"}, {"C", "int"}, {"D", "int"}, {"E", "int"},
		{"F", "int"}, {"G", "int"}, {"H", "int"}, {"I", "int"}, {"J", "int"},
		{"K", "int"}, {"L", "int"}, {"M", "int"}, {"N", "int"}, {"O", "int"},
		{"P", "int"}, {"Q", "int"}, {"R", "int"}, {"S", "int"}, {"T", "int"},
		{"U", "int"}, {"V", "int"}, {"W", "int"}, {"X", "int"}, {"Y", "int"}, {"Z", "int"},
		{"F1", "int"}, {"F2", "int"}, {"F3", "int"}, {"F4", "int"},
		{"F5", "int"}, {"F6", "int"}, {"F7", "int"}, {"F8", "int"},
		{"F9", "int"}, {"F10", "int"}, {"F11", "int"}, {"F12", "int"},
		{"LeftShift", "int"}, {"RightShift", "int"},
		{"LeftControl", "int"}, {"RightControl", "int"},
		{"LeftAlt", "int"}, {"RightAlt", "int"},
	},
	"Mouse": {
		{"Left", "int"}, {"Right", "int"}, {"Middle", "int"},
	},
	"Gamepad": {
		{"LeftStickX", "int"}, {"LeftStickY", "int"},
		{"RightStickX", "int"}, {"RightStickY", "int"},
		{"LeftTrigger", "int"}, {"RightTrigger", "int"},
	},
}

// Keywords for completion.
var Keywords = []string{
	"if", "else", "for", "while", "return", "with",
	"class", "enum", "event", "new", "fn",
	"true", "false", "this", "not", "and", "or", "is", "in",
}

// BuiltinTypes for type completion.
var BuiltinTypes = []string{
	"int", "float", "bool", "string",
	"vec2", "vec3", "vec4", "mat4", "quat",
	"List", "fn",
	"Color", "Rectangle", "Texture2D", "Font", "Sound",
	"Camera2D", "Camera3D",
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
	"Color": {
		{Label: "r", Detail: "int", Kind: CompletionKindField},
		{Label: "g", Detail: "int", Kind: CompletionKindField},
		{Label: "b", Detail: "int", Kind: CompletionKindField},
		{Label: "a", Detail: "int", Kind: CompletionKindField},
	},
	"Rectangle": {
		{Label: "x", Detail: "float", Kind: CompletionKindField},
		{Label: "y", Detail: "float", Kind: CompletionKindField},
		{Label: "width", Detail: "float", Kind: CompletionKindField},
		{Label: "height", Detail: "float", Kind: CompletionKindField},
	},
}

// BuiltinTypeFieldTypes maps built-in type fields to their types (for chained resolution).
var BuiltinTypeFieldTypes = map[string]map[string]string{
	"vec2":      {"x": "float", "y": "float"},
	"vec3":      {"x": "float", "y": "float", "z": "float"},
	"vec4":      {"x": "float", "y": "float", "z": "float", "w": "float"},
	"quat":      {"x": "float", "y": "float", "z": "float", "w": "float"},
	"Color":     {"r": "int", "g": "int", "b": "int", "a": "int"},
	"Rectangle": {"x": "float", "y": "float", "width": "float", "height": "float"},
}

// ModuleNames for completion.
var ModuleNames = []string{"math", "io", "rl"}

// NamespaceNames for completion.
var NamespaceNames = []string{"Colors", "Key", "Mouse", "Gamepad"}
