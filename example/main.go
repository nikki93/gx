package main

func fib(n int) int {
	if n <= 1 {
		return n
	} else {
		return fib(n-1) + fib(n-2)
	}
}

func testFib() {
	assert(fib(6) == 8)
}

func testUnary() {
	assert(-(3) == -3)
	assert(+(3) == 3)
}

func testVariables() {
	x := 3
	y := 4
	assert(x == 3)
	assert(y == 4)
	y = y + 2
	x = x + 1
	assert(y == 6)
	assert(x == 4)
}

func main() {
	testFib()
	testUnary()
	testVariables()
}
