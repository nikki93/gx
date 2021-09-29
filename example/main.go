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
	check(fib(6) == 8)
}

func testUnary() {
	check(-(3) == -3)
	check(+(3) == 3)
}

func testVariables() {
	x := 3
	y := 4
	check(x == 3)
	check(y == 4)
	y = y + 2
	x = x + 1
	check(y == 6)
	check(x == 4)
	y += 2
	x += 1
	check(y == 8)
	check(x == 5)
}

func testIf() {
	x := 0
	if cond := false; cond {
		x = 2
	}
	check(x == 0)
}

func testFor() {
	{
		sum := 0
		for i := 0; i < 5; i += 1 {
			sum += i
		}
		check(sum == 10)
	}
	{
		sum := 0
		i := 0
		for i < 5 {
			sum += i
			i += 1
		}
		check(sum == 10)
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
		check(sum == 10)
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
	check(val == 42)
	ptr := &val
	*ptr = 14
	check(val == 14)
	setToFortyTwo(ptr)
	check(val == 42)
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
		check(s.x == 0)
		check(s.y == 0)
		check(s.inner.z == 0)
		{
			p := &s
			p.x = 2
			check(p.x == 2)
			check(s.x == 2)
			s.y = 4
			check(p.y == 4)
		}
		check(outerSum(s) == 6)
		setXToFortyTwo(&s)
		check(s.x == 42)
	}
	{
		s := Outer{2, 3, Inner{4}}
		check(s.x == 2)
		check(s.y == 3)
		check(s.inner.z == 4)
		s.x += 1
		s.y += 1
		s.inner.z += 1
		check(s.x == 3)
		check(s.y == 4)
		check(s.inner.z == 5)
	}
	{
		s := Outer{x: 2, y: 3, inner: Inner{z: 4}}
		check(s.x == 2)
		check(s.y == 3)
		check(s.inner.z == 4)
	}
	{
		s := Outer{
			x: 2,
			y: 3,
			inner: Inner{
				z: 4,
			},
		}
		check(s.x == 2)
		check(s.y == 3)
		check(s.inner.z == 4)
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
		check(i == 14)
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
	check(p.sum() == 5)
	ptr := &p
	check(ptr.sum() == 5) // Pointer as value receiver
	p.setZero()           // Addressable value as pointer receiver
	check(p.x == 0)
	check(p.y == 0)
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

type Holder[T any] struct {
	item T
}

func incrHolder[T Numeric](h *Holder[T]) {
	h.item += 1
}

func (h Holder[T]) get() T {
	return h.item
}

func (h *Holder[T]) set(item T) {
	h.item = item
}

func testGenerics() {
	{
		check(add(1, 2) == 3)
		check(add(1.2, 2.0) == 3.2)
	}
	{
		i := Holder[int]{42}
		check(i.item == 42)
		incrHolder(&i)
		check(i.item == 43)

		f := Holder[float64]{42}
		check(f.item == 42)
		check(add(f.item, 20) == 62)
		incrHolder(&f)
		check(f.item == 43)

		p := Holder[Point]{Point{1, 2}}
		check(p.item.x == 1)
		check(p.item.y == 2)
		p.item.setZero()
		check(p.item.x == 0)
		check(p.item.y == 0)

		p.set(Point{3, 2})
		check(p.item.x == 3)
		check(p.item.y == 2)
		check(p.get().x == 3)
		check(p.get().y == 2)
	}
}

//
// Lambdas
//

func iterateOneToTen(f func(int)) {
	for i := 1; i <= 10; i += 1 {
		f(i)
	}
}

func testLambdas() {
	{
		val := 42
		check(val == 42)
		foo := func(newVal int) {
			val = newVal
		}
		foo(14)
		check(val == 14)

		val2 := func() int {
			return val
		}()
		check(val2 == val)
	}
	{
		sum := 0
		iterateOneToTen(func(i int) {
			sum += i
		})
		check(sum == 55)
	}
}

//
// Arrays
//

func setSecondElementToThree(arr *[4]int) {
	arr[1] = 3
}

type HasArray struct {
	arr [4]int
}

func testArrays() {
	{
		arr := [4]int{1, 2, 3, 4}
		check(arr[2] == 3)
		sum := 0
		for i := 0; i < len(arr); i += 1 {
			sum += arr[i]
		}
		check(sum == 10)
		check(arr[1] == 2)
		setSecondElementToThree(&arr)
		check(arr[1] == 3)
	}
	{
		arr := [...]int{1, 2, 3, 4, 5}
		check(len(arr) == 5)
	}
	{
		arr := [...][2]int{{1, 2}, {3, 4}}
		check(len(arr) == 2)
		check(arr[0][0] == 1)
		check(arr[0][1] == 2)
		check(arr[1][0] == 3)
		check(arr[1][1] == 4)
	}
	{
		h := HasArray{[4]int{1, 2, 3, 4}}
		check(len(h.arr) == 4)
		check(h.arr[2] == 3)
	}
}

//
// Slices
//

func appendFortyTwo(s *[]int) {
	*s = append(*s, 42)
}

func testSlices() {
	{
		s := []int{}
		check(len(s) == 0)
		s = append(s, 1)
		s = append(s, 2)
		check(len(s) == 2)
		check(s[0] == 1)
		check(s[1] == 2)
		appendFortyTwo(&s)
		check(len(s) == 3)
		check(s[2] == 42)
	}
	{
		s := []int{1, 2, 3}
		check(len(s) == 3)
	}
}

//
// Seq
//

type Seq[T any] []T

func (s *Seq[T]) Add(val T) {
	*s = append(*s, val)
}

func testSeqs() {
	{
		s := Seq[int]{}
		check(len(s) == 0)
		s.Add(1)
		s.Add(2)
		check(len(s) == 2)
		check(s[0] == 1)
		check(s[1] == 2)
	}
	{
		s := Seq[int]{1, 2, 3}
		check(len(s) == 3)
	}
	{
		s := Seq[Point]{{1, 2}, {3, 4}}
		check(len(s) == 2)
		check(s[0].x == 1)
		check(s[0].y == 2)
		check(s[1].x == 3)
		check(s[1].y == 4)
		s.Add(Point{5, 6})
		check(s[2].x == 5)
		check(s[2].y == 6)
	}
	{
		s := Seq[Seq[int]]{{1}, {}, {3, 4}}
		check(len(s) == 3)
		check(len(s[0]) == 1)
		check(s[0][0] == 1)
		check(len(s[1]) == 0)
		check(len(s[2]) == 2)
		check(s[2][0] == 3)
		check(s[2][1] == 4)
	}
}

//
// Main
//

func main() {
	testFib()
	testUnary()
	testVariables()
	testIf()
	testFor()
	testPointer()
	testStruct()
	testMethod()
	testGenerics()
	testLambdas()
	testArrays()
	testSlices()
	testSeqs()
}
