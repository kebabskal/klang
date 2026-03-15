Item:class {
	name:string
	cost:int
}

ConsumableItem:Item {
	consume() {
		print("Consumed: ", name)
	}
}

items:List<Item> = [
	{name = "sword", cost = 10},
	{name = "shield", cost = 45},
]

main() {
	potion:ConsumableItem = {name = "potion", cost = 8}
	potion.cost = 666
	items.append(potion)

	potion.consume()
}
