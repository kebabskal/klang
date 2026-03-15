generic_function<T>(message:string, argument:T) {
	print(message, " => ", argument)
}

GenericClass:class<T> {
	value:T

	new(num:T) {
		value = num
	}
}

main() {
	generic_function<int>("test explicit", 5)
	generic_function("test implicit", "implicit")

	explicit_instance: = GenericClass<int>(10)
	print("explicit_instance: ", explicit_instance.value)
	
	implicit_instance: = GenericClass("value")
	print("implicit_instance: ", implicit_instance.value)
}