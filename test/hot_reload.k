
i := 0

test() {
	print("A: ", i)
	i += 1
}

main() {
	i = 0
	
	while true {
		test()
		wait(5.0)
	}
}