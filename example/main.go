package main

//
// Basics
//

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

func testIf() {
	x := 0
	if cond := false; cond {
		x = 2
	}
	assert(x == 0)
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

//
// Pointers
//

func setToFortyTwo(ptr *int) {
	*ptr = 42
}

func testPointer() {
	val := 42
	assert(val == 42)
	ptr := &val
	*ptr = 14
	assert(val == 14)
	setToFortyTwo(ptr)
	assert(val == 42)
}

//
// Structs
//

type Outer struct {
	x     int
	y     int
	inner Inner
}

type Inner struct {
	z int
}

func outerSum(o Outer) int {
	return o.x + o.y + o.inner.z
}

func setXToFortyTwo(o *Outer) {
	o.x = 42
}

type PtrPtr struct {
	pp **int // Should be formatted as `int **pp;` in C++
}

func testStruct() {
	{
		s := Outer{}
		assert(s.x == 0)
		assert(s.y == 0)
		assert(s.inner.z == 0)
		{
			p := &s
			p.x = 2
			assert(p.x == 2)
			assert(s.x == 2)
			s.y = 4
			assert(p.y == 4)
		}
		assert(outerSum(s) == 6)
		setXToFortyTwo(&s)
		assert(s.x == 42)
	}
	{
		s := Outer{2, 3, Inner{4}}
		assert(s.x == 2)
		assert(s.y == 3)
		assert(s.inner.z == 4)
		s.x += 1
		s.y += 1
		s.inner.z += 1
		assert(s.x == 3)
		assert(s.y == 4)
		assert(s.inner.z == 5)
	}
	{
		s := Outer{x: 2, y: 3, inner: Inner{z: 4}}
		assert(s.x == 2)
		assert(s.y == 3)
		assert(s.inner.z == 4)
	}
	{
		s := Outer{
			x: 2,
			y: 3,
			inner: Inner{
				z: 4,
			},
		}
		assert(s.x == 2)
		assert(s.y == 3)
		assert(s.inner.z == 4)
	}
	{
		// Out-of-order elements in struct literal no longer allowed
		//s := Outer{
		//  inner: Inner{
		//    z: 4,
		//  },
		//  y: 3,
		//  x: 2,
		//}
	}
	{
		i := 42
		p := &i
		pp := &p
		d := PtrPtr{pp}
		**d.pp = 14
		assert(i == 14)
	}
}

//
// Methods
//

type Point struct {
	x, y float32
}

func (p Point) sum() float32 {
	return p.x + p.y
}

func (p *Point) setZero() {
	p.x = 0
	p.y = 0
}

func testMethod() {
	p := Point{2, 3}
	assert(p.sum() == 5)
	ptr := &p
	assert(ptr.sum() == 5) // Pointer as value receiver
	p.setZero()            // Addressable value as pointer receiver
	assert(p.x == 0)
	assert(p.y == 0)
}

//
// Generics
//

type Numeric interface {
	int | float64
}

func add[T Numeric](a, b T) T {
	return a + b
}

type Item[T any] struct {
	item T
}

func testGenerics() {
	{
		assert(add(1, 2) == 3)
		assert(add(1.2, 2.0) == 3.2)
	}
	{
		i := Item[int]{42}
		assert(i.item == 42)
		f := Item[float64]{42}
		assert(f.item == 42)
	}
}

//
// Main
//

func main() {
	testFib()
	testUnary()
	testVariables()
	testFor()
	testPointer()
	testStruct()
	testMethod()
	testGenerics()
}
