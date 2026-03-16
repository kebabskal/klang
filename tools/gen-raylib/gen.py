#!/usr/bin/env python3
"""
Raylib binding generator for Klang.

Reads raylib_api.json (from raylib's parser output) and generates:
  - libs/raylib/raylib.go   (analysis + codegen registration)
  - runtime/kl_raylib.h     (C wrapper functions for vec2/vec3 conversions)

Usage:
  python3 tools/gen-raylib/gen.py [path/to/raylib_api.json]

If no path is given, it tries /tmp/raylib_api.json.
"""

import json
import re
import sys
import os
from collections import OrderedDict

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

# Functions to skip entirely (internal, complex pointer APIs, etc.)
SKIP_FUNCTIONS = {
    # Text formatting (variadic)
    "TextFormat",
    # rlgl / low-level
    "rlglInit", "rlglClose",
    # Callbacks
    "SetTraceLogCallback", "SetLoadFileDataCallback", "SetSaveFileDataCallback",
    "SetLoadFileTextCallback", "SetSaveFileTextCallback",
    "AttachAudioMixedProcessor", "DetachAudioMixedProcessor",
    "AttachAudioStreamProcessor", "DetachAudioStreamProcessor",
    # Memory management
    "MemAlloc", "MemRealloc", "MemFree",
    # File data (raw bytes)
    "LoadFileData", "UnloadFileData", "SaveFileData",
    "LoadFileText", "UnloadFileText", "SaveFileText",
    # Directory/path utilities (return complex types)
    "LoadDirectoryFiles", "UnloadDirectoryFiles",
    "LoadDroppedFiles", "UnloadDroppedFiles",
    # Compression
    "CompressData", "DecompressData", "EncodeDataBase64", "DecodeDataBase64",
    # Automation events
    "LoadAutomationEventList", "UnloadAutomationEventList",
    "SetAutomationEventList", "SetAutomationEventBaseFrame",
    "StartAutomationEventRecording", "StopAutomationEventRecording",
    "PlayAutomationEvent", "ExportAutomationEventList",
    # Image pixel manipulation (raw pointer APIs)
    "LoadImageColors", "LoadImagePalette", "UnloadImageColors", "UnloadImagePalette",
    "GetImageColor",
    # Image/Texture raw data
    "GetPixelDataSize", "LoadTextureFromImage", "LoadTextureCubemap",
    "UpdateTexture", "UpdateTextureRec",
    # Font glyph data
    "LoadFontData", "UnloadFontData", "GenImageFontAtlas",
    "LoadFontEx", "LoadFontFromMemory",
    # Code points
    "LoadCodepoints", "UnloadCodepoints", "GetCodepointCount",
    "GetCodepoint", "GetCodepointNext", "GetCodepointPrevious",
    "CodepointToUTF8",
    # Mesh generation internals
    "UploadMesh", "UpdateMeshBuffer", "UnloadMesh",
    "GetMeshBoundingBox", "GenMeshTangents",
    "ExportMesh", "ExportMeshAsCode",
    # Material internals
    "LoadMaterials", "SetMaterialTexture", "SetModelMeshMaterial",
    "LoadMaterialDefault",
    # Model animation internals
    "LoadModelAnimations", "UnloadModelAnimations", "UnloadModelAnimation",
    "UpdateModelAnimation", "UpdateModelAnimationBones", "IsModelAnimationValid",
    # Shader internals
    "GetShaderLocation", "GetShaderLocationAttrib",
    "SetShaderValue", "SetShaderValueV", "SetShaderValueMatrix",
    "SetShaderValueTexture",
    # VR
    "LoadVrStereoConfig", "UnloadVrStereoConfig",
    # Audio stream low-level
    "LoadAudioStream", "UnloadAudioStream", "IsAudioStreamReady",
    "UpdateAudioStream", "IsAudioStreamProcessed",
    "PlayAudioStream", "PauseAudioStream", "ResumeAudioStream",
    "StopAudioStream", "IsAudioStreamPlaying",
    "SetAudioStreamVolume", "SetAudioStreamPitch", "SetAudioStreamPan",
    "SetAudioStreamBufferSizeDefault", "SetAudioStreamCallback",
    # Wave data manipulation
    "LoadWaveSamples", "UnloadWaveSamples",
    # Text internal
    "TextToInteger", "TextToFloat",
    "TextCopy", "TextIsEqual", "TextLength",
    "TextSubtext", "TextReplace", "TextInsert",
    "TextJoin", "TextSplit", "TextAppend",
    "TextFindIndex", "TextToUpper", "TextToLower", "TextToPascal",
    "TextToSnake", "TextToCamel",
    # GetRandom (Klang has its own random)
    "GetRandomValue", "SetRandomSeed", "LoadRandomSequence", "UnloadRandomSequence",
    # Spline internals
    "GetSplinePointLinear", "GetSplinePointBasis", "GetSplinePointCatmullRom",
    "GetSplinePointBezierQuad", "GetSplinePointBezierCubic",
    # Misc internals
    "OpenURL", "TraceLog", "SetTraceLogLevel",
    "TakeScreenshot", "ExportImage", "ExportImageAsCode", "ExportImageToMemory",
    # Pointer-heavy draw functions
    "DrawSplineBasis", "DrawSplineBezierQuadratic", "DrawSplineBezierCubic",
    "DrawSplineLinear", "DrawSplineCatmullRom", "DrawSplineSegmentLinear",
    "DrawSplineSegmentBasis", "DrawSplineSegmentCatmullRom",
    "DrawSplineSegmentBezierQuadratic", "DrawSplineSegmentBezierCubic",
    # Draw mesh / material (complex pointer APIs)
    "DrawMesh", "DrawMeshInstanced",
    # Texture filter/wrap (use enums namespace instead)
    "GenTextureMipmaps",
    # Wave manipulation (modifies in place)
    "WaveFormat", "WaveCopy", "WaveCrop",
    # Check collision point-to-poly (takes pointer array)
    "CheckCollisionPointPoly",
    # GetGlyphInfo etc
    "GetGlyphIndex", "GetGlyphInfo", "GetGlyphAtlasRec",
    # Image gen from raw data
    "LoadImageRaw", "LoadImageSvg", "LoadImageAnimFromMemory",
    "LoadImageFromMemory", "LoadImageFromTexture", "LoadImageFromScreen",
}

# Param types that make a function un-bindable
SKIP_PARAM_TYPES = {
    "...",
    "AudioCallback",
    "TraceLogCallback",
    "LoadFileDataCallback",
    "SaveFileDataCallback",
    "LoadFileTextCallback",
    "SaveFileTextCallback",
    "unsigned char *",
    "const unsigned char *",
    "void *",
    "const void *",
    "int *",
    "unsigned int *",
    "float *",
    "char **",
    "const Vector2 *",
    "const Vector3 *",
    "const float *",
    "const int *",
    "Vector2 *",
    "Vector3 *",
    "const char **",
    "GlyphInfo *",
    "Rectangle *",
    "Rectangle **",
    "const Rectangle *",
    "Matrix *",
    "const Matrix *",
    "Material *",
    "MaterialMap *",
    "char *",
    "FilePathList",
    "AutomationEventList",
    "AutomationEventList *",
    "AutomationEvent",
}

# Return types that make a function un-bindable
SKIP_RETURN_TYPES = {
    "void *",
    "unsigned char *",
    "char *",
    "const char **",
    "int *",
    "float *",
    "FilePathList",
    "AutomationEventList",
}

# C type → Klang type mapping
TYPE_MAP = {
    "void": "void",
    "int": "int",
    "float": "float",
    "double": "float",
    "bool": "bool",
    "unsigned int": "int",
    "unsigned char": "int",
    "long": "int",
    "const char *": "string",
    "Vector2": "vec2",
    "Vector3": "vec3",
    "Vector4": "vec4",
    "Quaternion": "vec4",
    "Color": "Color",
    "Rectangle": "Rectangle",
    "Texture2D": "Texture2D",
    "Texture": "Texture2D",
    "RenderTexture2D": "RenderTexture",
    "RenderTexture": "RenderTexture",
    "Camera2D": "Camera2D",
    "Camera3D": "Camera3D",
    "Camera": "Camera3D",
    "Image": "Image",
    "Font": "Font",
    "Sound": "Sound",
    "Music": "Music",
    "Model": "Model",
    "Mesh": "Mesh",
    "Shader": "Shader",
    "Ray": "Ray",
    "RayCollision": "RayCollision",
    "BoundingBox": "BoundingBox",
    "Material": "Material",
    "Wave": "Wave",
    "ModelAnimation": "ModelAnimation",
    "Transform": "Transform",
    "BoneInfo": "BoneInfo",
    "NPatchInfo": "NPatchInfo",
    "Matrix": "Matrix",
    "GlyphInfo": "GlyphInfo",
    "AudioStream": "AudioStream",
}

# Types that need vec2/vec3 wrappers (Klang native ↔ raylib C)
VECTOR_TYPES = {"Vector2", "Vector3", "Vector4", "Quaternion"}

# Structs to expose as Klang types with field completions
EXPOSED_STRUCTS = {
    "Color": {"r": "int", "g": "int", "b": "int", "a": "int"},
    "Rectangle": {"x": "float", "y": "float", "width": "float", "height": "float"},
    "Image": {"width": "int", "height": "int"},
    "Texture": {"id": "int", "width": "int", "height": "int"},
    "Camera2D": {"offset": "vec2", "target": "vec2", "rotation": "float", "zoom": "float"},
    "Camera3D": {"position": "vec3", "target": "vec3", "up": "vec3", "fovy": "float"},
    "Ray": {"position": "vec3", "direction": "vec3"},
    "RayCollision": {"hit": "bool", "distance": "float", "point": "vec3", "normal": "vec3"},
    "BoundingBox": {"min": "vec3", "max": "vec3"},
    "Model": {"transform": "Matrix", "meshCount": "int", "materialCount": "int"},
    "Font": {"baseSize": "int"},
    "Sound": {},
    "Music": {"looping": "bool"},
    "Wave": {"frameCount": "int", "sampleRate": "int", "sampleSize": "int", "channels": "int"},
    "Shader": {"id": "int"},
    "NPatchInfo": {"left": "int", "top": "int", "right": "int", "bottom": "int"},
    "RenderTexture": {"id": "int"},
}

# Types that use struct literal constructors (passed as {field, field, ...})
CONSTRUCTOR_TYPES = {
    "Color": [("r", "int"), ("g", "int"), ("b", "int"), ("a", "int")],
    "Rectangle": [("x", "float"), ("y", "float"), ("w", "float"), ("h", "float")],
}

# Types passed by value in C (not as pointers)
VALUE_TYPES = [
    "Color", "Rectangle", "Camera2D", "Camera3D",
    "Texture2D", "Texture", "RenderTexture2D", "RenderTexture",
    "Font", "Sound", "Music", "Image",
    "Model", "Mesh", "Shader", "Material",
    "Ray", "RayCollision", "BoundingBox",
    "Wave", "ModelAnimation", "Transform",
    "NPatchInfo", "Matrix", "Vector2", "Vector3", "Vector4",
    "GlyphInfo", "BoneInfo", "AudioStream",
]

# Enum name → Klang namespace name
ENUM_NAMESPACE_MAP = {
    "ConfigFlags": "Flag",
    "KeyboardKey": "Key",
    "MouseButton": "Mouse",
    "MouseCursor": "Cursor",
    "GamepadButton": "GamepadButton",
    "GamepadAxis": "Gamepad",
    "CameraMode": "CameraMode",
    "CameraProjection": "CameraProjection",
    "BlendMode": "BlendMode",
    "TextureFilter": "TextureFilter",
    "TextureWrap": "TextureWrap",
    "Gesture": "Gesture",
    "TraceLogLevel": "LogLevel",
    "MaterialMapIndex": "MaterialMap",
    "PixelFormat": "PixelFormat",
    "FontType": "FontType",
    "CubemapLayout": "CubemapLayout",
    "NPatchLayout": "NPatchLayout",
    "ShaderLocationIndex": "ShaderLoc",
    "ShaderUniformDataType": "ShaderUniform",
}

# Enum value prefix stripping rules
ENUM_PREFIX_STRIP = {
    "ConfigFlags": "FLAG_",
    "KeyboardKey": "KEY_",
    "MouseButton": "MOUSE_BUTTON_",
    "MouseCursor": "MOUSE_CURSOR_",
    "GamepadButton": "GAMEPAD_BUTTON_",
    "GamepadAxis": "GAMEPAD_AXIS_",
    "CameraMode": "CAMERA_",
    "CameraProjection": "CAMERA_",
    "BlendMode": "BLEND_",
    "TextureFilter": "TEXTURE_FILTER_",
    "TextureWrap": "TEXTURE_WRAP_",
    "Gesture": "GESTURE_",
    "TraceLogLevel": "LOG_",
    "MaterialMapIndex": "MATERIAL_MAP_",
    "PixelFormat": "PIXELFORMAT_",
    "FontType": "FONT_",
    "CubemapLayout": "CUBEMAP_LAYOUT_",
    "NPatchLayout": "NPATCH_",
    "ShaderLocationIndex": "SHADER_LOC_",
    "ShaderUniformDataType": "SHADER_UNIFORM_",
}

# Manual overrides: klang_name → C function name (for functions that need wrappers or special handling)
MANUAL_OVERRIDES = {
    # Helpers that already exist
    "color": "kl_color",
    "color_rgb": "kl_color_rgb",
    "rect": "kl_rect",
    "camera2d": "kl_camera2d",
    "camera3d": "kl_camera3d",
}

# Manual additions: functions not in the API but useful as helpers
MANUAL_FUNCTIONS = [
    # Helper constructors
    {
        "name": "color",
        "sig": "color(r:int, g:int, b:int, a:int):Color",
        "c_func": "kl_color",
        "return_type": "Color",
    },
    {
        "name": "color_rgb",
        "sig": "color_rgb(r:int, g:int, b:int):Color",
        "c_func": "kl_color_rgb",
        "return_type": "Color",
    },
    {
        "name": "rect",
        "sig": "rect(x:float, y:float, w:float, h:float):Rectangle",
        "c_func": "kl_rect",
        "return_type": "Rectangle",
    },
    {
        "name": "camera2d",
        "sig": "camera2d(offset:vec2, target:vec2, rotation:float, zoom:float):Camera2D",
        "c_func": "kl_camera2d",
        "return_type": "Camera2D",
    },
    {
        "name": "camera3d",
        "sig": "camera3d(position:vec3, target:vec3, up:vec3, fovy:float):Camera3D",
        "c_func": "kl_camera3d",
        "return_type": "Camera3D",
    },
    {
        "name": "get_screen_size",
        "sig": "get_screen_size():vec2",
        "c_func": "kl_get_screen_size",
        "return_type": "vec2",
    },
]

# Functions that modify via pointer (Image*, Camera*, etc.) - we handle specific ones
MUTABLE_PTR_OVERRIDES = {
    # UpdateCamera takes Camera* - we wrap it to take/return by value
    "UpdateCamera": {
        "skip": True,  # Too complex for auto-gen, would need special wrapper
    },
}

# Functions where we override parameter names for clarity
PARAM_NAME_OVERRIDES = {}


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def pascal_to_snake(name: str) -> str:
    """Convert PascalCase to snake_case, handling numbers and acronyms.

    Special handling: keeps digit+letter combos together when they form
    well-known suffixes like 2D, 3D, 2d, 3d, etc.
    """
    # Insert underscore before uppercase letters preceded by lowercase
    result = re.sub(r'([a-z])([A-Z])', r'\1_\2', name)
    # Insert underscore before uppercase letters followed by lowercase, preceded by uppercase
    result = re.sub(r'([A-Z]+)([A-Z][a-z])', r'\1_\2', result)
    # Insert underscore before numbers preceded by letters (but not within known combos)
    result = re.sub(r'([a-zA-Z])(\d)', r'\1_\2', result)
    # Insert underscore after numbers followed by letters
    result = re.sub(r'(\d)([a-zA-Z])', r'\1_\2', result)
    result = result.lower()
    # Fix common suffixes: _2_d → _2d, _3_d → _3d, _4_x → _4x
    result = re.sub(r'_(\d)_([a-z])\b', r'_\1\2', result)
    # Fix v_1, v_2, v_3 → v1, v2, v3 (vertex names)
    result = re.sub(r'\bv_(\d)', r'v\1', result)
    return result


def klang_type(c_type: str):
    """Map a C type to a Klang type. Returns None if unmappable."""
    # Strip const and pointer for struct types passed by const ref
    clean = c_type.strip()
    if clean.startswith("const ") and clean.endswith(" *"):
        inner = clean[6:-2].strip()
        if inner in TYPE_MAP:
            return TYPE_MAP[inner]
    return TYPE_MAP.get(clean)


def needs_wrapper(func_data: dict) -> bool:
    """Check if a function needs a C wrapper (uses Vector2/Vector3 params or returns)."""
    ret = func_data.get("returnType", "void")
    if ret in VECTOR_TYPES:
        return True
    for p in func_data.get("params", []):
        ptype = p["type"].replace("const ", "").replace(" *", "").strip()
        if ptype in VECTOR_TYPES:
            return True
    return False


def is_bindable(func_data: dict) -> bool:
    """Check if a function can be bound to Klang."""
    name = func_data["name"]
    if name in SKIP_FUNCTIONS:
        return False
    if name in MUTABLE_PTR_OVERRIDES and MUTABLE_PTR_OVERRIDES[name].get("skip"):
        return False

    # Skip functions with un-bindable return types
    ret = func_data.get("returnType", "void")
    if ret in SKIP_RETURN_TYPES:
        return False
    if klang_type(ret) is None and ret != "void":
        return False

    # Skip functions with un-bindable param types
    for p in func_data.get("params", []):
        ptype = p["type"]
        if ptype in SKIP_PARAM_TYPES:
            return False
        # Mutable pointer to struct (Image*, Model*, etc.) - skip
        if ptype.endswith(" *") and not ptype.startswith("const "):
            inner = ptype[:-2].strip()
            if inner in TYPE_MAP and inner not in ("char",):
                return False
        if klang_type(ptype) is None:
            return False

    return True


def make_signature(klang_name: str, params: list, return_type: str) -> str:
    """Build a Klang signature string like 'func_name(p1:type1, p2:type2):ReturnType'."""
    param_strs = []
    for p in params:
        ktype = klang_type(p["type"])
        param_strs.append(f"{pascal_to_snake(p['name'])}:{ktype}")
    sig = f"{klang_name}({', '.join(param_strs)})"
    if return_type != "void":
        kret = klang_type(return_type)
        if kret and kret != "void":
            sig += f":{kret}"
    return sig


def c_param_type(ptype: str) -> str:
    """Get the C parameter type for wrapper generation."""
    clean = ptype.strip()
    # const Struct * → Struct (pass by value in wrapper)
    if clean.startswith("const ") and clean.endswith(" *"):
        inner = clean[6:-2].strip()
        if inner in TYPE_MAP:
            return inner
    return clean


def enum_value_to_klang(value_name: str, prefix: str) -> str:
    """Convert a C enum value name to a Klang-style PascalCase name."""
    stripped = value_name
    if prefix and stripped.startswith(prefix):
        stripped = stripped[len(prefix):]

    # Convert SCREAMING_SNAKE to PascalCase
    parts = stripped.split("_")
    result = ""
    for part in parts:
        if part:
            result += part[0].upper() + part[1:].lower()
    return result


# ---------------------------------------------------------------------------
# Generator
# ---------------------------------------------------------------------------

class RaylibGenerator:
    def __init__(self, api_data: dict):
        self.api = api_data
        self.functions = []  # List of {klang_name, sig, c_func, return_type, needs_wrapper, params, original_name}
        self.namespaces = OrderedDict()  # ns_name → [(klang_member, c_value, detail)]
        self.wrappers = []  # C wrapper functions to generate

    def process(self):
        self._process_functions()
        self._process_enums()
        self._process_color_defines()
        self._add_manual_functions()

    def _process_functions(self):
        for func in self.api["functions"]:
            if not is_bindable(func):
                continue

            name = func["name"]
            klang_name = pascal_to_snake(name)
            params = func.get("params", [])
            ret = func.get("returnType", "void")

            # Build signature
            sig = make_signature(klang_name, params, ret)

            # Determine C function to call
            if needs_wrapper(func):
                c_func = f"kl_{klang_name}"
                self._generate_wrapper(func, c_func)
            else:
                c_func = name

            # Determine return type for codegen
            kret = klang_type(ret) if ret != "void" else None

            self.functions.append({
                "klang_name": klang_name,
                "sig": sig,
                "c_func": c_func,
                "return_type": kret,
                "original_name": name,
            })

    def _generate_wrapper(self, func: dict, wrapper_name: str):
        """Generate a C wrapper function for vec2/vec3 conversions."""
        name = func["name"]
        params = func.get("params", [])
        ret = func.get("returnType", "void")

        # Build wrapper parameter list
        wrapper_params = []
        call_args = []

        for p in params:
            ptype = c_param_type(p["type"])
            pname = p["name"]

            if ptype == "Vector2":
                wrapper_params.append(f"vec2 {pname}")
                call_args.append(f"(Vector2){{{pname}.x, {pname}.y}}")
            elif ptype == "Vector3":
                wrapper_params.append(f"vec3 {pname}")
                call_args.append(f"(Vector3){{{pname}.x, {pname}.y, {pname}.z}}")
            elif ptype == "Vector4" or ptype == "Quaternion":
                wrapper_params.append(f"vec4 {pname}")
                call_args.append(f"(Vector4){{{pname}.x, {pname}.y, {pname}.z, {pname}.w}}")
            else:
                wrapper_params.append(f"{ptype} {pname}")
                call_args.append(pname)

        param_str = ", ".join(wrapper_params) if wrapper_params else "void"
        args_str = ", ".join(call_args)

        # Return type conversion
        if ret == "Vector2":
            ret_c = "vec2"
            call = f"Vector2 _r = {name}({args_str}); return (vec2){{ _r.x, _r.y }}"
        elif ret == "Vector3":
            ret_c = "vec3"
            call = f"Vector3 _r = {name}({args_str}); return (vec3){{ _r.x, _r.y, _r.z }}"
        elif ret == "Vector4" or ret == "Quaternion":
            ret_c = "vec4"
            call = f"Vector4 _r = {name}({args_str}); return (vec4){{ _r.x, _r.y, _r.z, _r.w }}"
        elif ret == "void":
            ret_c = "void"
            call = f"{name}({args_str})"
        else:
            ret_c = ret
            call = f"return {name}({args_str})"

        wrapper = f"static inline {ret_c} {wrapper_name}({param_str}) {{ {call}; }}"
        self.wrappers.append(wrapper)

    def _process_enums(self):
        for enum in self.api["enums"]:
            ename = enum["name"]
            if ename not in ENUM_NAMESPACE_MAP:
                continue

            ns_name = ENUM_NAMESPACE_MAP[ename]
            prefix = ENUM_PREFIX_STRIP.get(ename, "")
            members = []

            for val in enum.get("values", []):
                c_name = val["name"]
                klang_name = enum_value_to_klang(c_name, prefix)
                detail = "int"
                members.append((klang_name, c_name, detail))

            self.namespaces[ns_name] = members

    def _process_color_defines(self):
        """Extract color #define constants into the Colors namespace."""
        # Special name mapping for colors that are single words (no underscore)
        COLOR_NAME_MAP = {
            "LIGHTGRAY": "LightGray",
            "DARKGRAY": "DarkGray",
            "DARKGREEN": "DarkGreen",
            "SKYBLUE": "SkyBlue",
            "DARKBLUE": "DarkBlue",
            "DARKPURPLE": "DarkPurple",
            "DARKBROWN": "DarkBrown",
            "RAYWHITE": "RayWhite",
        }
        members = []
        for define in self.api.get("defines", []):
            if define.get("type") == "COLOR":
                c_name = define["name"]
                klang_name = COLOR_NAME_MAP.get(c_name, enum_value_to_klang(c_name, ""))
                members.append((klang_name, c_name, "Color"))
        if members:
            self.namespaces["Colors"] = members

    def _add_manual_functions(self):
        # Add manual helper functions (avoid duplicates)
        existing = {f["klang_name"] for f in self.functions}
        for mf in MANUAL_FUNCTIONS:
            if mf["name"] not in existing:
                self.functions.append({
                    "klang_name": mf["name"],
                    "sig": mf["sig"],
                    "c_func": mf["c_func"],
                    "return_type": mf["return_type"],
                    "original_name": mf["name"],
                })

    # -----------------------------------------------------------------------
    # Output: libs/raylib/raylib.go
    # -----------------------------------------------------------------------

    def generate_go(self) -> str:
        lines = []
        lines.append('// Code generated by tools/gen-raylib/gen.py. DO NOT EDIT.')
        lines.append('//')
        lines.append('// To regenerate, run:')
        lines.append('//   python3 tools/gen-raylib/gen.py path/to/raylib_api.json')
        lines.append('package raylib')
        lines.append('')
        lines.append('import (')
        lines.append('\t"github.com/klang-lang/klang/internal/analysis"')
        lines.append('\t"github.com/klang-lang/klang/internal/codegen"')
        lines.append(')')
        lines.append('')
        lines.append('func init() {')

        # --- analysis.RegisterVendor ---
        lines.append('\tanalysis.RegisterVendor(&analysis.VendorLib{')

        # Modules
        lines.append('\t\tModules: map[string][]analysis.StdlibFunc{')
        lines.append('\t\t\t"rl": {')
        # Group functions by category
        categories = self._categorize_functions()
        for cat, funcs in categories.items():
            lines.append(f'\t\t\t\t// {cat}')
            for f in funcs:
                lines.append(f'\t\t\t\t{{"{f["klang_name"]}", "{f["sig"]}"}},')
        lines.append('\t\t\t},')
        lines.append('\t\t},')

        # Namespaces
        lines.append('\t\tNamespaces: map[string][]analysis.StdlibFunc{')
        for ns_name, members in self.namespaces.items():
            lines.append(f'\t\t\t"{ns_name}": {{')
            for klang_name, _, detail in members:
                lines.append(f'\t\t\t\t{{"{klang_name}", "{detail}"}},')
            lines.append('\t\t\t},')
        lines.append('\t\t},')

        # Types
        type_names = list(EXPOSED_STRUCTS.keys())
        # Map "Texture" to "Texture2D" in Klang
        klang_type_names = []
        for t in type_names:
            if t == "Texture":
                klang_type_names.append("Texture2D")
            else:
                klang_type_names.append(t)
        lines.append('\t\tTypes: []string{')
        for t in klang_type_names:
            lines.append(f'\t\t\t"{t}",')
        lines.append('\t\t},')

        # TypeMembers
        lines.append('\t\tTypeMembers: map[string][]analysis.CompletionItem{')
        for struct_name, fields in EXPOSED_STRUCTS.items():
            if not fields:
                continue
            display_name = "Texture2D" if struct_name == "Texture" else struct_name
            lines.append(f'\t\t\t"{display_name}": {{')
            for field_name, field_type in fields.items():
                lines.append(f'\t\t\t\t{{Label: "{field_name}", Detail: "{field_type}", Kind: analysis.CompletionKindField}},')
            lines.append('\t\t\t},')
        lines.append('\t\t},')

        # TypeFieldTypes
        lines.append('\t\tTypeFieldTypes: map[string]map[string]string{')
        for struct_name, fields in EXPOSED_STRUCTS.items():
            if not fields:
                continue
            display_name = "Texture2D" if struct_name == "Texture" else struct_name
            field_map = ", ".join(f'"{k}": "{v}"' for k, v in fields.items())
            lines.append(f'\t\t\t"{display_name}": {{{field_map}}},')
        lines.append('\t\t},')

        # BuiltinConstructors
        lines.append('\t\tBuiltinConstructors: map[string]string{')
        for type_name, params in CONSTRUCTOR_TYPES.items():
            param_str = ", ".join(f"{p[0]}:{p[1]}" for p in params)
            lines.append(f'\t\t\t"{type_name}": "{type_name}({param_str})",')
        lines.append('\t\t},')

        # BuiltinConstructorParams
        lines.append('\t\tBuiltinConstructorParams: map[string][]string{')
        for type_name, params in CONSTRUCTOR_TYPES.items():
            types_str = ", ".join(f'"{p[1]}"' for p in params)
            lines.append(f'\t\t\t"{type_name}": {{{types_str}}},')
        lines.append('\t\t},')

        # CTypeMap
        lines.append('\t\tCTypeMap: map[string]string{')
        for struct_name in EXPOSED_STRUCTS:
            display_name = "Texture2D" if struct_name == "Texture" else struct_name
            c_name = "Texture2D" if struct_name == "Texture" else struct_name
            lines.append(f'\t\t\t"{display_name}": "{c_name}",')
        lines.append('\t\t},')

        # BuiltinIdents
        all_idents = list(klang_type_names) + ["rl"] + list(self.namespaces.keys())
        lines.append('\t\tBuiltinIdents: []string{')
        for ident in all_idents:
            lines.append(f'\t\t\t"{ident}",')
        lines.append('\t\t},')

        # ModuleNamespaces — map "rl" → all enum namespace names
        lines.append('\t\tModuleNamespaces: map[string][]string{')
        lines.append('\t\t\t"rl": {')
        for ns_name in self.namespaces:
            lines.append(f'\t\t\t\t"{ns_name}",')
        lines.append('\t\t\t},')
        lines.append('\t\t},')

        lines.append('\t})')
        lines.append('')

        # --- codegen.RegisterVendorCodegen ---
        lines.append('\tcodegen.RegisterVendorCodegen(&codegen.VendorCodegen{')

        # ModuleFuncs
        lines.append('\t\tModuleFuncs: map[string]map[string]string{')
        lines.append('\t\t\t"rl": {')
        for cat, funcs in categories.items():
            lines.append(f'\t\t\t\t// {cat}')
            for f in funcs:
                lines.append(f'\t\t\t\t"{f["klang_name"]}": "{f["c_func"]}",')
        lines.append('\t\t\t},')
        lines.append('\t\t},')

        # ConstNamespaces
        lines.append('\t\tConstNamespaces: map[string]map[string]string{')
        for ns_name, members in self.namespaces.items():
            lines.append(f'\t\t\t"{ns_name}": {{')
            for klang_name, c_name, _ in members:
                lines.append(f'\t\t\t\t"{klang_name}": "{c_name}",')
            lines.append('\t\t\t},')
        lines.append('\t\t},')

        # ValueTypes
        lines.append('\t\tValueTypes: []string{')
        for vt in VALUE_TYPES:
            lines.append(f'\t\t\t"{vt}",')
        lines.append('\t\t},')

        # TypeMap
        lines.append('\t\tTypeMap: map[string]string{')
        for struct_name in EXPOSED_STRUCTS:
            display_name = "Texture2D" if struct_name == "Texture" else struct_name
            c_name = "Texture2D" if struct_name == "Texture" else struct_name
            lines.append(f'\t\t\t"{display_name}": "{c_name}",')
        # Also add "Texture" alias
        lines.append(f'\t\t\t"Texture": "Texture2D",')
        lines.append('\t\t},')

        # ConstructorTypes
        lines.append('\t\tConstructorTypes: []string{')
        for ct in CONSTRUCTOR_TYPES:
            lines.append(f'\t\t\t"{ct}",')
        lines.append('\t\t},')

        # ReturnTypes
        lines.append('\t\tReturnTypes: map[string]string{')
        for f in self.functions:
            if f["return_type"]:
                lines.append(f'\t\t\t"{f["klang_name"]}": "{f["return_type"]}",')
        lines.append('\t\t},')

        # DetectIdents
        detect_idents = ["rl"] + list(self.namespaces.keys())
        lines.append(f'\t\tDetectIdents: []string{{')
        for di in detect_idents:
            lines.append(f'\t\t\t"{di}",')
        lines.append('\t\t},')
        lines.append('\t\tDetectModule: "rl",')

        lines.append('\t})')
        lines.append('}')
        lines.append('')

        return "\n".join(lines)

    def _categorize_functions(self) -> OrderedDict:
        """Group functions by category based on naming patterns."""
        categories = OrderedDict()
        cat_patterns = [
            ("Window", ["InitWindow", "CloseWindow", "WindowShouldClose", "IsWindow", "SetWindow",
                        "GetScreen", "ToggleFullscreen", "ToggleBorderless", "MaximizeWindow",
                        "MinimizeWindow", "RestoreWindow", "SetWindowIcon", "SetWindowIcons",
                        "GetWindow", "GetMonitor", "SetWindowMin", "SetWindowMax",
                        "SetWindowPosition", "SetWindowOpacity", "SetWindowFocused",
                        "GetCurrentMonitor", "GetRenderWidth", "GetRenderHeight",
                        "EnableEventWaiting", "DisableEventWaiting",
                        "IsWindowState", "SetWindowState", "ClearWindowState"]),
            ("Timing", ["SetTargetFPS", "GetFrameTime", "GetTime", "GetFPS", "WaitTime"]),
            ("Cursor", ["ShowCursor", "HideCursor", "IsCursorHidden", "EnableCursor",
                        "DisableCursor", "IsCursorOnScreen", "SetMouseCursor"]),
            ("Drawing", ["BeginDrawing", "EndDrawing", "ClearBackground",
                         "BeginMode2D", "EndMode2D", "BeginMode3D", "EndMode3D",
                         "BeginTextureMode", "EndTextureMode",
                         "BeginShaderMode", "EndShaderMode",
                         "BeginBlendMode", "EndBlendMode",
                         "BeginScissorMode", "EndScissorMode"]),
            ("Shapes", ["DrawPixel", "DrawLine", "DrawCircle", "DrawEllipse",
                        "DrawRing", "DrawRectangle", "DrawTriangle", "DrawPoly",
                        "DrawSector", "DrawCapsule"]),
            ("3D Shapes", ["DrawLine3D", "DrawPoint3D", "DrawCircle3D",
                           "DrawTriangle3D", "DrawCube", "DrawSphere",
                           "DrawCylinder", "DrawCapsule3D", "DrawPlane", "DrawGrid",
                           "DrawRay", "DrawBillboard"]),
            ("Textures", ["LoadTexture", "UnloadTexture", "DrawTexture",
                          "LoadRenderTexture", "UnloadRenderTexture",
                          "IsTextureValid", "IsRenderTextureValid",
                          "SetTextureFilter", "SetTextureWrap",
                          "GenTextureMipmaps"]),
            ("Image", ["LoadImage", "UnloadImage", "IsImageValid",
                       "GenImage", "ImageCopy", "ImageFromImage", "ImageFromChannel",
                       "ImageFormat", "ImageToPOT",
                       "ImageCrop", "ImageAlpha", "ImageResize",
                       "ImageFlip", "ImageRotate", "ImageMipmaps",
                       "ImageDither", "ImageKernelConvolution",
                       "ImageText", "ImageDraw", "ImageClear",
                       "ImageColorTint", "ImageColorInvert", "ImageColorGrayscale",
                       "ImageColorContrast", "ImageColorBrightness", "ImageColorReplace"]),
            ("Text", ["DrawText", "MeasureText", "LoadFont", "UnloadFont",
                      "IsFontValid", "LoadFontFromImage", "MeasureTextEx",
                      "GetFontDefault", "DrawTextPro"]),
            ("Models", ["LoadModel", "UnloadModel", "IsModelValid",
                        "DrawModel", "DrawModelWires",
                        "DrawBoundingBox",
                        "GenMesh", "LoadModelFromMesh"]),
            ("Input — Keyboard", ["IsKeyPressed", "IsKeyDown", "IsKeyReleased", "IsKeyUp",
                                  "IsKeyPressedRepeat", "GetKeyPressed", "GetCharPressed",
                                  "SetExitKey"]),
            ("Input — Mouse", ["IsMouseButton", "GetMouse", "SetMouse"]),
            ("Input — Gamepad", ["IsGamepad", "GetGamepad", "SetGamepad", "GetGamepadName"]),
            ("Input — Touch", ["GetTouch", "SetGestures", "IsGesture", "GetGesture"]),
            ("Audio", ["InitAudioDevice", "CloseAudioDevice", "IsAudioDeviceReady",
                       "SetMasterVolume", "GetMasterVolume"]),
            ("Sound", ["LoadSound", "UnloadSound", "PlaySound", "StopSound",
                       "PauseSound", "ResumeSound", "IsSoundPlaying", "IsSoundValid",
                       "SetSoundVolume", "SetSoundPitch", "SetSoundPan",
                       "LoadSoundAlias", "UnloadSoundAlias", "LoadSoundFromWave"]),
            ("Music", ["LoadMusicStream", "UnloadMusicStream", "PlayMusicStream",
                       "StopMusicStream", "PauseMusicStream", "ResumeMusicStream",
                       "UpdateMusicStream", "IsMusicStreamPlaying", "IsMusicValid",
                       "SetMusicVolume", "SetMusicPitch", "SetMusicPan",
                       "GetMusicTimeLength", "GetMusicTimePlayed",
                       "SeekMusicStream", "SetMusicLooping"]),
            ("Wave", ["LoadWave", "UnloadWave", "IsWaveValid", "ExportWave"]),
            ("Collision", ["CheckCollision"]),
            ("Camera", ["UpdateCamera", "GetCamera", "GetWorldToScreen",
                        "GetScreenToWorld"]),
            ("Shader", ["LoadShader", "UnloadShader", "IsShaderValid",
                        "GetShaderDefault"]),
            ("Misc", ["DrawFPS", "SetExitKey", "SetConfigFlags",
                      "GetColor", "ColorToInt", "ColorToHSV",
                      "ColorFromHSV", "ColorBrightness", "ColorContrast",
                      "ColorAlpha", "ColorAlphaBlend", "ColorLerp",
                      "Fade", "SetRandomSeed"]),
        ]

        categorized = set()
        for cat, prefixes in cat_patterns:
            cat_funcs = []
            for f in self.functions:
                if f["original_name"] in categorized:
                    continue
                for prefix in prefixes:
                    if f["original_name"].startswith(prefix) or f["original_name"] == prefix:
                        cat_funcs.append(f)
                        categorized.add(f["original_name"])
                        break
            if cat_funcs:
                categories[cat] = cat_funcs

        # Uncategorized
        remaining = [f for f in self.functions if f["original_name"] not in categorized]
        if remaining:
            categories["Other"] = remaining

        return categories

    # -----------------------------------------------------------------------
    # Output: runtime/kl_raylib.h
    # -----------------------------------------------------------------------

    def generate_header(self) -> str:
        lines = []
        lines.append('// Code generated by tools/gen-raylib/gen.py. DO NOT EDIT.')
        lines.append('//')
        lines.append('// To regenerate, run:')
        lines.append('//   python3 tools/gen-raylib/gen.py path/to/raylib_api.json')
        lines.append('')
        lines.append('#ifndef KL_RAYLIB_H')
        lines.append('#define KL_RAYLIB_H')
        lines.append('')
        lines.append('#include "raylib.h"')
        lines.append('')
        lines.append('// --- Manual helpers ---')
        lines.append('')
        lines.append('static inline Color kl_color(int r, int g, int b, int a) {')
        lines.append('    return (Color){ (unsigned char)r, (unsigned char)g, (unsigned char)b, (unsigned char)a };')
        lines.append('}')
        lines.append('')
        lines.append('static inline Color kl_color_rgb(int r, int g, int b) {')
        lines.append('    return (Color){ (unsigned char)r, (unsigned char)g, (unsigned char)b, 255 };')
        lines.append('}')
        lines.append('')
        lines.append('static inline Camera2D kl_camera2d(vec2 offset, vec2 target, float rotation, float zoom) {')
        lines.append('    Camera2D cam = {0};')
        lines.append('    cam.offset = (Vector2){ offset.x, offset.y };')
        lines.append('    cam.target = (Vector2){ target.x, target.y };')
        lines.append('    cam.rotation = rotation;')
        lines.append('    cam.zoom = zoom;')
        lines.append('    return cam;')
        lines.append('}')
        lines.append('')
        lines.append('static inline Camera3D kl_camera3d(vec3 position, vec3 target, vec3 up, float fovy, int projection) {')
        lines.append('    Camera3D cam = {0};')
        lines.append('    cam.position = (Vector3){ position.x, position.y, position.z };')
        lines.append('    cam.target = (Vector3){ target.x, target.y, target.z };')
        lines.append('    cam.up = (Vector3){ up.x, up.y, up.z };')
        lines.append('    cam.fovy = fovy;')
        lines.append('    cam.projection = projection;')
        lines.append('    return cam;')
        lines.append('}')
        lines.append('')
        lines.append('static inline Rectangle kl_rect(float x, float y, float w, float h) {')
        lines.append('    return (Rectangle){ x, y, w, h };')
        lines.append('}')
        lines.append('')
        lines.append('static inline vec2 kl_get_screen_size(void) {')
        lines.append('    return (vec2){ (float)GetScreenWidth(), (float)GetScreenHeight() };')
        lines.append('}')
        lines.append('')
        lines.append('// --- Auto-generated wrappers (vec2/vec3 conversions) ---')
        lines.append('')

        # Deduplicate wrappers (manual helpers are already above)
        manual_names = {
            "kl_color", "kl_color_rgb", "kl_camera2d", "kl_camera3d",
            "kl_rect", "kl_get_screen_size",
        }
        for wrapper in self.wrappers:
            # Extract wrapper name
            match = re.search(r'\b(kl_\w+)\s*\(', wrapper)
            if match and match.group(1) in manual_names:
                continue
            lines.append(wrapper)
            lines.append('')

        lines.append('#endif // KL_RAYLIB_H')
        lines.append('')
        return "\n".join(lines)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    api_path = sys.argv[1] if len(sys.argv) > 1 else "/tmp/raylib_api.json"
    if not os.path.exists(api_path):
        print(f"Error: {api_path} not found", file=sys.stderr)
        print("Download it from: https://raw.githubusercontent.com/raysan5/raylib/5.5/parser/output/raylib_api.json", file=sys.stderr)
        sys.exit(1)

    with open(api_path) as f:
        api = json.load(f)

    gen = RaylibGenerator(api)
    gen.process()

    # Determine output paths relative to this script
    script_dir = os.path.dirname(os.path.abspath(__file__))
    project_root = os.path.dirname(os.path.dirname(script_dir))

    go_path = os.path.join(project_root, "libs", "raylib", "raylib.go")
    h_path = os.path.join(project_root, "runtime", "kl_raylib.h")

    go_code = gen.generate_go()
    h_code = gen.generate_header()

    with open(go_path, "w") as f:
        f.write(go_code)
    print(f"Generated {go_path} ({len(gen.functions)} functions)")

    with open(h_path, "w") as f:
        f.write(h_code)
    print(f"Generated {h_path} ({len(gen.wrappers)} wrappers)")

    # Summary
    total_ns = sum(len(v) for v in gen.namespaces.values())
    print(f"Namespaces: {len(gen.namespaces)} ({total_ns} total constants)")
    print(f"Types: {len(EXPOSED_STRUCTS)}")


if __name__ == "__main__":
    main()
