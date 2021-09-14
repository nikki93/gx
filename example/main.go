package example

func add(x, y int) int {
	return x + y
}

func main() {
	print("1 == 1: ", 1 == 1, "\n")
	print("1 == 2: ", 1 == 2, "\n")
	print("2 + 3: ", add(2, getThree()), "\n")
}
