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

func main() {
	testFib()
	testUnary()
}
