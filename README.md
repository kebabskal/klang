# Klang

An experimental programming language. Syntax designed by @kebabskal, implemented by Claude Code (Opus 4.6) in under two days with some handholding.

The language itself is inspired by GDScript, C# and Odin.

Surprisingly it both kinda works, and is pretty smooth to develop with, due to a significant percentage (probably a majority) of the development time being spent creating an LSP and VSCode integration.

performance isn't too shabby since

Goes without saying, but don't use this for anything serious (or anything, really). Been a fun experiment!

<video src="https://github.com/kebabskal/klang/raw/master/demo-720.mov" controls width="100%"></video>

*examples/raylib/main.k*

Check [`language_tour.k`](examples/language_tour.k) for a comprehensive tour of the language.

```
Main:class

main() {
    print("hello world")
}
```

## Features

- **Compiles to C** — generates readable C code, compiles with any C compiler
- **Hot reload** — `kl dev` watches files and hot-reloads code changes without restarting
- **Classes & inheritance** — full OOP with constructors, methods, properties, and events
- **Generics** — type-safe generic classes and methods with monomorphization
- **Lambdas & closures** — first-class functions with variable capture
- **Built-in collections** — `List<T>` and `Dictionary<K, V>` with functional methods
- **Vector math** — built-in `vec2`, `vec3`, `vec4`, `quat` types with operator overloading
- **Standard library** — `math`, `io`, and `Random` modules
- **Type casting** — `int()`, `float()`, `bool()`, `string()` conversions
- **VS Code extension** — syntax highlighting, completions, hover, signature help, go-to-definition
- **Raylib integration** — optional vendor lib for game development

## Quick Start

### Build from source

```sh
git clone https://github.com/klang-lang/klang
cd klang
make build
```

Requires Go 1.26+ and a C compiler (`cc`, `gcc`, or `clang`).

This produces `bin/kl` and `bin/kl-lsp`.

### Install VS Code extension

```sh
make install-lsp
```

Then restart VS Code or run "Restart Language Server".

## CLI

```sh
kl run main.k              # Build and run
kl build main.k             # Build only (output: build/game)
kl build main.k release     # Optimized release build
kl dev main.k               # Watch & hot reload
kl lsp                      # Start language server
```

All commands accept multiple files or directories:

```sh
kl run main.k entity.k health.k
kl run examples/entities/
```

## Language Guide

### Variables

```
x := 42                   # type inferred as int
pi := 3.14                # float
name := "hello"           # string
alive := true             # bool

score:int = 100            # explicit type annotation
```

### Functions

```
add(a:int, b:int):int {
    return a + b
}

greet(name:string) {
    print("Hello, ", name)
}
```

### Control Flow

```
if score >= 90 {
    print("A")
} else if score >= 80 {
    print("B")
} else {
    print("F")
}

for i in 10 {              # 0 to 9
    print(i)
}

for item in list {          # iterate collection
    print(item)
}

for key, value in dict {    # key-value iteration
    print(key, " = ", value)
}

while alive {
    update()
}
```

### Operators

```
# Arithmetic: + - * / %
# Compound:   += -= *= /=
# Comparison: == != < > <= >=
# Logical:    and  or  not
```

### Classes

```
Animal:class {
    name:string
    sound:string = "..."
    legs:int = 4

    new(name:string) {
        this.name = name
    }

    speak() {
        print(name, " says ", sound)
    }
}

# Inheritance
Dog:Animal {
    breed:string = "unknown"

    fetch(item:string) {
        print(name, " fetches the ", item)
    }
}
```

Create instances:

```
dog := Animal("Rex")
dog.sound = "woof"

# Struct-like initialization
bird:Animal = {name = "Tweety", sound = "tweet", legs = 2}
```

### Properties

```
Counter:class {
    value:int = 0
    max:int = 100

    # Read-only computed property
    percentage:int {
        get => value * 100 / max
    }

    # Read-write property
    is_maxed:bool {
        get => value >= max
        set(v) => {
            if v { value = max }
        }
    }
}
```

### Events

```
Health:class {
    amount:int = 100
    died:event(delta:int)

    damage(n:int) {
        amount -= n
        if amount <= 0 {
            died.emit(n)
        }
    }
}

# Usage
h := Health()
h.died.connect((delta) => print("died from ", delta, " damage"))
h.damage(150)
```

### Generics

```
Stack:class<T> {
    items:List<T> = []

    push(item:T) {
        items.append(item)
    }

    pop():T {
        return items.pop()
    }
}

stack := Stack<string>()
stack.push("hello")
```

### Enumerations

```
Season:enum = {
    Spring = 0,
    Summer,
    Autumn,
    Winter,
}
```

### Lambdas & Closures

```
numbers.filter((x) => x > 10)
numbers.sort((a, b) => a - b)
items.find((x) => x.name == "target")

# Stored closures
callback := fn {
    print("called!")
}
callback()
```

### Collections

**List\<T\>**

```
items:List<int> = []
items.append(42)
items.insert(0, 10)
items.remove(0)
items.pop()
items.count()
items.first()
items.last()
items.contains(42)
items.index_of(42)
items.reverse()
items.clone()
items.slice(0, 3)
items.clear()

# Functional
items.filter((x) => x > 10)
items.map((x) => x * 2)
items.find((x) => x == 42)
items.find_index((x) => x == 42)
items.remove_all((x) => x < 0)
items.sort((a, b) => a - b)
items.sort_by((x) => x.score)
```

**Dictionary\<K, V\>**

```
ages:Dictionary<string, int> = {
    "Alice": 30,
    "Bob": 25,
}

ages.append("Charlie", 35)
ages.set("Alice", 31)
ages.get("Alice")
ages.has("Bob")
ages.remove("Bob")
ages.keys()
ages.values()
ages.count()
ages.clear()

value := ages["Alice"]
```

### Type Casting

```
f := 3.14
i := int(f)          # 3
g := float(42)       # 42.0
s := string(i)       # "3"
b := bool(1)         # true
```

### Vector Types

```
pos := vec2(1.0, 2.0)
dir := vec3(0.0, 1.0, 0.0)
color := vec4(1.0, 0.0, 0.0, 1.0)
rot := quat(0.0, 0.0, 0.0, 1.0)

# Arithmetic works on vectors
result := pos + vec2(3.0, 4.0)
```

### Modules

```
# math module
y := math.sin(math.PI / 2.0)
d := math.clamp(value, 0, 100)

# io module
io.write_file("out.txt", "hello")
content := io.read_file("out.txt")
exists := io.file_exists("out.txt")
io.delete_file("out.txt")

# 'with' brings module into scope
with math {
    print(sin(PI))
}
```

**math** — `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `atan2`, `sqrt`, `pow`, `abs`, `floor`, `ceil`, `round`, `min`, `max`, `clamp`, `lerp`, `sign`, `deg2rad`, `rad2deg`. Constants: `PI`, `TAU`, `E`, `INF`, `EPSILON`.

**io** — `read_file`, `write_file`, `append_file`, `file_exists`, `delete_file`, `create_dir`, `dir_exists`, `list_dir`.

**Random** — seeded PRNG: `Random(seed)` with `.rangei(min, max)`, `.float()`, `.bool()`.

### Inline C

For when you need to drop down to C:

```
@c {
printf("Hello from C!\n");
}
```

## Hot Reload

`kl dev` watches your source files and recompiles on save. If your main class has `render()` or `update()` methods, it uses DLL hot-reloading — your program keeps running and picks up code changes instantly. Otherwise it does a full restart.

```sh
kl dev main.k
```

## Raylib

Klang includes an optional [raylib](https://www.raylib.com/) vendor library for game development. Functions are available through the `rl` module:

```
Main:class

main() {
    rl.init_window(800, 600, "my game")
    rl.set_target_fps(60)

    while not rl.window_should_close() {
        with rl {
            begin_drawing()
            clear_background(color(32, 32, 32, 255))
            draw_text("hello!", 20, 20, 20, color(255, 255, 255, 255))
            end_drawing()
        }
    }

    rl.close_window()
}
```

Build with raylib installed on your system. Covers window management, drawing, input (keyboard, mouse, gamepad), textures, fonts, and audio.

## Examples

| Example                                                | Description                                             |
| ------------------------------------------------------ | ------------------------------------------------------- |
| [`examples/language_tour.k`](examples/language_tour.k) | Comprehensive tour of all language features             |
| [`examples/entities/`](examples/entities/)             | Multi-file project with classes, events, and properties |
| [`examples/raylib/`](examples/raylib/)                 | Raylib graphics demo                                    |

Run any example:

```sh
kl run examples/language_tour.k
kl run examples/entities/
```

## Project Structure

```
cmd/kl/          CLI tool (build, run, dev, lsp)
cmd/kl-lsp/      Language server binary
internal/parser/ Lexer, AST, parser
internal/analysis/ Type checking, LSP features
internal/codegen/  C code generator
runtime/         C runtime headers
libs/raylib/     Raylib vendor library
editors/vscode-klang/ VS Code extension
examples/        Example programs
```

## License

MIT
