Main:class

some_event:event(value:int)

on_some_event(value:int) {
  print("on_some_event(", value, ")")
}

main() {
  some_event.connect(on_some_event)

  some_event.emit(1337)
  # some_event.emit("this should not work")  # type error: string not assignable to int

  some_event.disconnect(on_some_event)
  some_event.emit(666)

  print("nothing printed right?")

}