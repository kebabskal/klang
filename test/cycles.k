Game:class

Entity:class {
	name:string
	health:int = 100
	target:Entity

	attack() {
		if target is Entity {
			print(name, " attacks ", target.name)
		}
	}
}

main() {
	a:Entity = {name = "Knight", health = 100}
	b:Entity = {name = "Dragon", health = 200}

	# create a cycle: a -> b -> a
	a.target = b
	b.target = a

	a.attack()
	b.attack()

	print("done")
}
