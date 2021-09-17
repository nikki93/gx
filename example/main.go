package main

type Struct struct {
	x int
	y int
	i InnerStruct
}

type InnerStruct struct {
	z int
}

func testStruct() {
	{
		s := Struct{}
		assert(s.x == 0)
		assert(s.y == 0)
		assert(s.i.z == 0)
	}
	{
		s := Struct{2, 3, InnerStruct{4}}
		assert(s.x == 2)
		assert(s.y == 3)
		assert(s.i.z == 4)
		s.x += 1
		s.y += 1
		s.i.z += 1
		assert(s.x == 3)
		assert(s.y == 4)
		assert(s.i.z == 5)
	}
	{
		s := Struct{x: 2, y: 3, i: InnerStruct{z: 4}}
		assert(s.x == 2)
		assert(s.y == 3)
		assert(s.i.z == 4)
	}
	{
		s := Struct{i: InnerStruct{z: 4}, y: 3, x: 2}
		assert(s.x == 2)
		assert(s.y == 3)
		assert(s.i.z == 4)
	}
}

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
	y += 2
	x += 1
	assert(y == 8)
	assert(x == 5)
}

func testFor() {
	{
		sum := 0
		for i := 0; i < 5; i += 1 {
			sum += i
		}
		assert(sum == 10)
	}
	{
		sum := 0
		i := 0
		for i < 5 {
			sum += i
			i += 1
		}
		assert(sum == 10)
	}
	{
		sum := 0
		i := 0
		for {
			if i >= 5 {
				break
			}
			sum += i
			i += 1
		}
		assert(sum == 10)
	}
}

func main() {
	testFib()
	testUnary()
	testVariables()
	testFor()
	testStruct()
}
