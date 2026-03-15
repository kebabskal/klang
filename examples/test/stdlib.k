# Standard library test

main() {
    # --- Math module ---
    angle := 1.5708
    print("=== Math ===")
    print(math.sin(angle))
    print(math.cos(0.0))
    print(math.sqrt(16.0))
    print(math.floor(3.7))
    print(math.ceil(3.2))
    print(math.round(3.5))
    print(math.abs(-42))
    print(math.min(10, 20))
    print(math.max(10, 20))
    print(math.clamp(15, 0, 10))
    print(math.lerp(0.0, 100.0, 0.5))
    print(math.PI)
    print(math.sign(-5))

    # --- Math with 'with' keyword ---
    print("=== Math (with) ===")
    with math {
        print(sin(0.0))
        print(PI)
        print(lerp(0.0, 10.0, 0.25))
    }

    # --- Vector types ---
    print("=== Vectors ===")
    v2 := vec2(1.0, 2.0)
    print(v2.x, v2.y)

    v3 := vec3(1.0, 2.0, 3.0)
    print(v3.x, v3.y, v3.z)

    v4 := vec4(1.0, 2.0, 3.0, 4.0)
    print(v4.x, v4.y, v4.z, v4.w)

    # vec3 math
    a := vec3(1.0, 0.0, 0.0)
    b := vec3(0.0, 1.0, 0.0)
    c := vec3_cross(a, b)
    print("cross:", c.x, c.y, c.z)

    len := vec3_length(vec3(3.0, 4.0, 0.0))
    print("length:", len)

    d := vec3_dot(a, b)
    print("dot:", d)

    # --- Quaternion ---
    q := quat(0.0, 0.0, 0.0, 1.0)
    print("quat:", q.x, q.y, q.z, q.w)

    # --- IO module ---
    print("=== IO ===")
    io.write_file("build/test_out.txt", "hello from klang!")
    content := io.read_file("build/test_out.txt")
    print(content)
    print(io.file_exists("build/test_out.txt"))
    io.delete_file("build/test_out.txt")
    print(io.file_exists("build/test_out.txt"))

    # --- Random ---
    print("=== Random ===")
    rnd := Random(1234)
    print(rnd.rangei(0, 100))
    print(rnd.rangei(0, 100))
    print(rnd.rangei(0, 100))
    print(rnd.float())
    print(rnd.bool())

    # Deterministic: same seed → same sequence
    rnd2 := Random(1234)
    print(rnd2.rangei(0, 100))

    # --- Inline C ---
    print("=== Inline C ===")
    @c {
    printf("hello from inline C!\n");
    int x = 42;
    printf("x = %d\n", x);
    }

    print("done!")
}
