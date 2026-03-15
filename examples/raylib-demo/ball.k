Ball:class {
	position:vec2
	velocity:vec2
	radius:float = 10

	update(delta:float) {
		position += velocity * vec2(delta, delta)
	}
}