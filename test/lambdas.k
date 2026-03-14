
main() {
	lambdas:List<fn> = []
	for i in 5 {
		some_lambda := fn {
			print(i)
		}

		lambdas.append(some_lambda)
	}

	for lambda in lambdas {
		lambda()
	}
}