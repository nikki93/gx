package example

func add(x, y int) int {
	return x + y
}

func sideEffect() bool {
	print("side effect!\n")
	return true
}

func main() {
	if sideEffect(); true {
		print("true\n")
	}
	assert(1 == 1)
	assert(add(2, 3) == 5)
}
