Game:class

Item:class {
	name:string
	cost:int
}

ConsumableItem:Item {
	consume() {
		print("consumed: ", name)
	}
}

main() {
	sword:Item = {name = "sword", cost = 100}
	potion:ConsumableItem = {name = "potion", cost = 8}

	items:List<Item> = [
		sword,
		potion,
	]

	for item in items {
		if item is ConsumableItem {
			item.consume()
		} else {
			print(item.name, " is not consumable")
		}
	}
}
