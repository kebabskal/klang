add(a:int, b:int):int {
	return a + b
}

greet(who:string) {
	print("hello ", who)
}

main() {
	# variables and types
	x := 42
	name := "klang"
	pi := 3.14
	alive := true

	print(x)
	print(name)
	print(pi)
	print(alive)

	# math
	y := x + 8
	print(y)

	# multi-arg print
	print("x is ", x)
	print("pi is ", pi)

	# for loop
	for i in 5 {
		print(i)
	}

	# if / else if / else
	score := 85
	if score >= 90 {
		print("A")
	} else if score >= 80 {
		print("B")
	} else if score >= 70 {
		print("C")
	} else {
		print("F")
	}

	# while
	count := 3
	while count > 0 {
		print(count)
		count -= 1
	}

	# method calls
	result := add(10, 20)
	print(result)
	greet("world")

	# collections
	numbers:List<int> = [1, 2]
	numbers.append(3)
	for i in numbers {
		print(i)
	}
}
