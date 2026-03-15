Main:class

main() {
  entity := Entity()

  entity.health.before_change((info) => {
    info.delta = info.delta / 2
  })
  entity.health.died((info) => print("death!"))

  while entity.health.is_alive {
    entity.health.change(-10)
    print(entity.health.amount)
  }
}