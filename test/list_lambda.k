Item:class {
	value:int = 0
	name:string = ""
}

TestLambda:class

main() {
	items:List<Item> = []
	for i in 5 {
		item := Item()
		item.value = i * 10
		items.append(item)
	}
	print("initial count: ", items.count())

	# filter
	big := items.filter((x) => x.value > 20)
	print("filter >20 count: ", big.count())

	# map — map to ints
	# values := items.map((x) => x.value)

	# find
	found := items.find((x) => x.value == 30)
	print("found value=30: ", found.value)

	# find_index
	idx := items.find_index((x) => x.value == 20)
	print("find_index value=20: ", idx)

	# remove_all
	items.remove_all((x) => x.value < 20)
	print("after remove_all <20: ", items.count())

	# sort
	items.sort((a, b) => a.value - b.value)
	print("sorted first: ", items.first().value)
	print("sorted last: ", items.last().value)

	# sort_by
	items.sort_by((x) => x.value)
	print("sort_by first: ", items.first().value)
}
