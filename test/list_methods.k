TestLists:class

main() {
	items:List<int> = []
	for i in 10 {
		items.append(i)
	}
	print("count: ", items.count())
	print("first: ", items.first())
	print("last: ", items.last())

	# reverse
	items.reverse()
	print("after reverse, first: ", items.first())

	# clone
	cloned := items.clone()
	print("cloned count: ", cloned.count())

	# contains / index_of
	print("contains 5: ", items.contains(5))
	print("index_of 5: ", items.index_of(5))

	# remove
	items.remove(0)
	print("after remove(0), count: ", items.count())

	# pop
	last := items.pop()
	print("popped: ", last)
	print("after pop, count: ", items.count())

	# clear
	cloned.clear()
	print("cloned after clear: ", cloned.count())

	# insert
	items.insert(0, 99)
	print("after insert(0,99), first: ", items.first())

	# slice
	sliced := items.slice(0, 3)
	print("sliced count: ", sliced.count())
}
