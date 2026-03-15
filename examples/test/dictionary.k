Main:class

main() {
  dict:Dictionary<string, int> = {
    "kebab": 999,
    "sausage": 123,
  }

  dict.append("fruit", 666)

  print(dict["kebab"])
  print(dict["something missing"])

  if dict.has("kebab") {
    print("hurray!")
  }

  for key in dict.keys() {
    print(key)
  } 

  for value in dict.values() {
    print(value)
  } 

  for key, value in dict {
    print(key, " => ", value)
  }
}