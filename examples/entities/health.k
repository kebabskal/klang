Health:class

ChangeInfo:class {
  delta:int
  health:Health
}

amount:int = 100
max:int = 100

before_change:event(info:ChangeInfo)
after_change:event(info:ChangeInfo)
died:event(info:ChangeInfo)


is_alive:bool {
  get => amount > 0
  set(alive) => {
    if not alive and is_alive {
      change(-amount)
    }
  }
}

change(delta:int):ChangeInfo {
  if not is_alive {
    return
  }

  info := ChangeInfo()
  info.delta = delta
  info.health = this
  before_change.emit(info)

  was_alive := is_alive
  amount += info.delta

  after_change.emit(info)

  if not is_alive and was_alive {
    # died
    died.emit(info)
  }

  return info
}