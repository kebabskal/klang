# ============================================================
#  Klang Language Tour
#  A comprehensive showcase of the language's features.
# ============================================================

Tour:class

# --- Classes ---

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

# Inheritance: Dog extends Animal
Dog:Animal {
	breed:string = "unknown"

	fetch(item:string) {
		print(name, " fetches the ", item, "!")
	}
}

Cat:Animal {
	indoor:bool = true
}

# --- Properties ---

Counter:class {
	value:int = 0
	max:int = 100

	# Computed read-only property
	percentage:int {
		get => value * 100 / max
	}

	# Read-write property with getter and setter
	is_maxed:bool {
		get => value >= max
		set(maxed) => {
			if maxed {
				value = max
			}
		}
	}

	increment(amount:int) {
		if value + amount > max {
			value = max
		} else {
			value += amount
		}
	}
}

# --- Events ---

score_changed:event(old_score:int, new_score:int)

on_score_changed(old_score:int, new_score:int) {
	print("  Score: ", old_score, " -> ", new_score)
}

# --- Enumerations ---

Season:enum = {
	Spring = 0,
	Summer,
	Autumn,
	Winter,
}

# --- Helper class for lambda demos ---

Value:class {
	n:int = 0
}

# --- Generics ---

Pair:class<A, B> {
	first:A
	second:B

	new(first:A, second:B) {
		this.first = first
		this.second = second
	}
}

Stack:class<T> {
	items:List<T> = []

	push(item:T) {
		items.append(item)
	}

	pop():T {
		return items.pop()
	}

	peek():T {
		return items.last()
	}

	size():int {
		return items.count()
	}
}

# --- Free functions ---

add(a:int, b:int):int {
	return a + b
}

factorial(n:int):int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n - 1)
}

greet(who:string) {
	print("Hello, ", who, "!")
}

# ============================================================
#  Main
# ============================================================

main() {
	print("=== Klang Language Tour ===")
	print("")

	# ----------------------------------------------------------
	# 1. Variables & Type Inference
	# ----------------------------------------------------------
	print("--- 1. Variables & Type Inference ---")

	x := 42                   # inferred as int
	pi := 3.14159             # inferred as float
	name := "Klang"           # inferred as string
	alive := true             # inferred as bool

	# Explicit type annotation
	score:int = 100
	label:string = "points"

	print(name, " score: ", score, " ", label)
	print("pi = ", pi, ", alive = ", alive)

	# ----------------------------------------------------------
	# 2. Arithmetic & Operators
	# ----------------------------------------------------------
	print("")
	print("--- 2. Arithmetic & Operators ---")

	a := 10
	b := 3
	print(a, " + ", b, " = ", a + b)
	print(a, " - ", b, " = ", a - b)
	print(a, " * ", b, " = ", a * b)
	print(a, " / ", b, " = ", a / b)
	print(a, " % ", b, " = ", a % b)

	# Compound assignment
	a += 5
	a -= 2
	a *= 3
	print("after +=5, -=2, *=3: ", a)

	# Logical operators use words: and, or, not
	print("true and false = ", true and false)
	print("true or false  = ", true or false)
	print("not true       = ", not true)

	# Comparison
	print("10 > 3  = ", 10 > 3)
	print("10 == 3 = ", 10 == 3)
	print("10 != 3 = ", 10 != 3)

	# ----------------------------------------------------------
	# 3. Control Flow
	# ----------------------------------------------------------
	print("")
	print("--- 3. Control Flow ---")

	# if / else if / else
	grade := 85
	if grade >= 90 {
		print("Grade: A")
	} else if grade >= 80 {
		print("Grade: B")
	} else if grade >= 70 {
		print("Grade: C")
	} else {
		print("Grade: F")
	}

	# for loop with range (0 to n-1)
	print("counting:")
	for i in 5 {
		print("  ", i)
	}

	# for loop over a list
	fruits:List<string> = ["apple", "banana", "cherry"]
	for fruit in fruits {
		print("  fruit: ", fruit)
	}

	# while loop
	countdown := 3
	while countdown > 0 {
		print("  T-", countdown)
		countdown -= 1
	}
	print("  Liftoff!")

	# ----------------------------------------------------------
	# 4. Functions
	# ----------------------------------------------------------
	print("")
	print("--- 4. Functions ---")

	result := add(17, 25)
	print("add(17, 25) = ", result)

	print("factorial(6) = ", factorial(6))

	greet("World")

	# ----------------------------------------------------------
	# 5. Classes & Objects
	# ----------------------------------------------------------
	print("")
	print("--- 5. Classes & Objects ---")

	# Constructor with 'new'
	dog := Animal("Rex")
	dog.sound = "woof"
	dog.speak()

	# Inline object initialization (struct-like syntax)
	bird:Animal = {name = "Tweety", sound = "tweet", legs = 2}
	bird.speak()
	print(bird.name, " has ", bird.legs, " legs")

	# Inheritance — Dog extends Animal
	puppy:Dog = {name = "Buddy", sound = "bark", breed = "Golden Retriever"}
	print(puppy.name, " is a ", puppy.breed)
	puppy.fetch("ball")

	# ----------------------------------------------------------
	# 6. Properties
	# ----------------------------------------------------------
	print("")
	print("--- 6. Properties ---")

	counter := Counter()
	counter.increment(75)
	print("value: ", counter.value)
	print("percentage: ", counter.percentage)
	print("is_maxed: ", counter.is_maxed)

	counter.increment(50)
	print("after max: value=", counter.value, ", is_maxed=", counter.is_maxed)

	# ----------------------------------------------------------
	# 7. Events
	# ----------------------------------------------------------
	print("")
	print("--- 7. Events ---")

	# Connect a named handler
	score_changed.connect(on_score_changed)
	score_changed.emit(0, 100)
	score_changed.emit(100, 250)

	# Disconnect
	score_changed.disconnect(on_score_changed)
	score_changed.emit(250, 300)
	print("  (nothing — handler disconnected)")

	# Re-connect and emit again
	score_changed.connect(on_score_changed)
	score_changed.emit(300, 500)

	# ----------------------------------------------------------
	# 8. Lists
	# ----------------------------------------------------------
	print("")
	print("--- 8. Lists ---")

	numbers:List<int> = []
	for i in 10 {
		numbers.append(i)
	}
	print("count: ", numbers.count())
	print("first: ", numbers.first())
	print("last: ", numbers.last())

	# contains / index_of
	print("contains 5: ", numbers.contains(5))
	print("index_of 5: ", numbers.index_of(5))

	# reverse
	numbers.reverse()
	print("after reverse, first: ", numbers.first())

	# clone
	cloned := numbers.clone()
	print("cloned count: ", cloned.count())

	# remove / pop
	numbers.remove(0)
	print("after remove(0), count: ", numbers.count())

	last := numbers.pop()
	print("popped: ", last)

	# insert / slice
	numbers.insert(0, 99)
	print("after insert(0,99), first: ", numbers.first())

	sliced := numbers.slice(0, 3)
	print("sliced count: ", sliced.count())

	# clear
	cloned.clear()
	print("cloned after clear: ", cloned.count())

	# ----------------------------------------------------------
	# 9. Lambdas & Functional List Methods
	# ----------------------------------------------------------
	print("")
	print("--- 9. Lambdas & Functional Methods ---")

	# Using objects for functional methods (filter, find, sort)
	items:List<Value> = []
	for i in 5 {
		item := Value()
		item.n = i * 10
		items.append(item)
	}
	print("initial count: ", items.count())

	# filter — returns a new list
	big := items.filter((x) => x.n > 20)
	print("filter (n > 20) count: ", big.count())

	# find — returns first match
	found := items.find((x) => x.n == 30)
	print("find n==30: ", found.n)

	# find_index
	idx := items.find_index((x) => x.n == 20)
	print("find_index n==20: ", idx)

	# remove_all — mutates in place
	items.remove_all((x) => x.n < 20)
	print("after remove_all (n<20): ", items.count())

	# sort with custom comparator
	items.sort((a, b) => a.n - b.n)
	print("sorted first: ", items.first().n)
	print("sorted last: ", items.last().n)

	# sort_by — shorthand for single-field sort
	items.sort_by((x) => x.n)
	print("sort_by first: ", items.first().n)

	# Storing lambdas
	callbacks:List<fn> = []
	for i in 3 {
		captured := i
		callbacks.append(fn {
			print("  lambda captured: ", captured)
		})
	}
	for cb in callbacks {
		cb()
	}

	# ----------------------------------------------------------
	# 10. Dictionaries
	# ----------------------------------------------------------
	print("")
	print("--- 10. Dictionaries ---")

	ages:Dictionary<string, int> = {
		"Alice": 30,
		"Bob": 25,
	}
	ages.append("Charlie", 35)

	print("Alice is ", ages["Alice"])
	print("has Bob: ", ages.has("Bob"))

	for key in ages.keys() {
		print("  key: ", key)
	}

	for value in ages.values() {
		print("  value: ", value)
	}

	# Iterate key-value pairs
	for key, value in ages {
		print("  ", key, " => ", value)
	}

	# ----------------------------------------------------------
	# 11. Generics
	# ----------------------------------------------------------
	print("")
	print("--- 11. Generics ---")

	# Generic class with multiple type parameters
	pair := Pair<string, int>("answer", 42)
	print("pair: ", pair.first, " = ", pair.second)

	# Generic class with single type parameter
	stack := Stack<string>()
	stack.push("first")
	stack.push("second")
	stack.push("third")
	print("stack size: ", stack.size())
	print("stack peek: ", stack.peek())
	print("stack pop: ", stack.pop())
	print("stack size after pop: ", stack.size())

	# ----------------------------------------------------------
	# 12. Vectors
	# ----------------------------------------------------------
	print("")
	print("--- 12. Vectors ---")

	# vec2, vec3, vec4, quat — built-in vector types
	pos := vec2(3.0, 4.0)
	print("vec2: (", pos.x, ", ", pos.y, ")")

	dir := vec3(1.0, 2.0, 3.0)
	print("vec3: (", dir.x, ", ", dir.y, ", ", dir.z, ")")

	rgba := vec4(1.0, 0.0, 0.0, 1.0)
	print("vec4: (", rgba.x, ", ", rgba.y, ", ", rgba.z, ", ", rgba.w, ")")

	q := quat(0.0, 0.0, 0.0, 1.0)
	print("quat: (", q.x, ", ", q.y, ", ", q.z, ", ", q.w, ")")

	# Vector arithmetic
	v1 := vec2(1.0, 2.0)
	v2 := vec2(3.0, 4.0)
	v3:vec2 = v1 + v2
	print("vec2 add: (", v3.x, ", ", v3.y, ")")

	# ----------------------------------------------------------
	# 13. Math Module & 'with' Statement
	# ----------------------------------------------------------
	print("")
	print("--- 13. Math Module ---")

	print("sin(PI/2):      ", math.sin(1.5708))
	print("cos(0):         ", math.cos(0.0))
	print("sqrt(144):      ", math.sqrt(144.0))
	print("floor(3.7):     ", math.floor(3.7))
	print("ceil(3.2):      ", math.ceil(3.2))
	print("abs(-42):       ", math.abs(-42))
	print("min(10, 20):    ", math.min(10, 20))
	print("max(10, 20):    ", math.max(10, 20))
	print("clamp(15,0,10): ", math.clamp(15, 0, 10))
	print("lerp(0,100,.5): ", math.lerp(0.0, 100.0, 0.5))
	print("sign(-5):       ", math.sign(-5))
	print("PI:             ", math.PI)

	# 'with' brings module members into scope
	with math {
		print("with math { sin(0) = ", sin(0.0), " }")
		print("with math { PI = ", PI, " }")
	}

	# ----------------------------------------------------------
	# 14. IO Module
	# ----------------------------------------------------------
	print("")
	print("--- 14. IO Module ---")

	io.write_file("build/tour_test.txt", "Hello from Klang!")
	content := io.read_file("build/tour_test.txt")
	print("file content: ", content)
	print("file exists: ", io.file_exists("build/tour_test.txt"))
	io.delete_file("build/tour_test.txt")
	print("after delete: ", io.file_exists("build/tour_test.txt"))

	# ----------------------------------------------------------
	# 15. Random
	# ----------------------------------------------------------
	print("")
	print("--- 15. Random ---")

	rng := Random(42)
	print("rangei(0,100): ", rng.rangei(0, 100))
	print("float():       ", rng.float())
	print("bool():        ", rng.bool())

	# Same seed = same sequence (deterministic)
	rng2 := Random(42)
	print("same seed:     ", rng2.rangei(0, 100))

	# ----------------------------------------------------------
	# 16. Inline C
	# ----------------------------------------------------------
	print("")
	print("--- 16. Inline C ---")

	@c {
	printf("  Hello from inline C!\n");
	int answer = 6 * 7;
	printf("  The answer is %d\n", answer);
	}

	# ----------------------------------------------------------
	print("")
	print("=== Tour Complete! ===")
}
