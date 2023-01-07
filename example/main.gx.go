//gx:include <string.h>
//gx:include "rect.hh"
//gx:include "sum_fields.hh"

package main

import (
	"github.com/nikki93/gx/example/foo"
	"github.com/nikki93/gx/example/person"
)

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

func testIncDec() {
	x := 0
	x++
	check(x == 1)
	x--
	check(x == 0)
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
		for i := 0; i < 5; i++ {
			sum += i
		}
		check(sum == 10)
	}
	{
		sum := 0
		i := 0
		for i < 5 {
			sum += i
			i++
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
			i++
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
		check(d.pp != nil)
		check(i == 14)
	}
	{
		p := PtrPtr{}
		check(p.pp == nil)
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
	Item T
}

func incrHolder[T Numeric](h *Holder[T]) {
	h.Item += 1
}

func (h Holder[T]) get() T {
	return h.Item
}

func (h *Holder[T]) set(Item T) {
	h.Item = Item
}

func testGenerics() {
	{
		check(add(1, 2) == 3)
		check(add(1.2, 2.0) == 3.2)
		check(add[float64](1.2, 2.0) == 3.2)
	}
	{
		i := Holder[int]{42}
		check(i.Item == 42)
		incrHolder(&i)
		check(i.Item == 43)

		f := Holder[float64]{42}
		check(f.Item == 42)
		check(add(f.Item, 20) == 62)
		incrHolder(&f)
		check(f.Item == 43)

		p := Holder[Point]{Point{1, 2}}
		check(p.Item.x == 1)
		check(p.Item.y == 2)
		p.Item.setZero()
		check(p.Item.x == 0)
		check(p.Item.y == 0)

		p.set(Point{3, 2})
		check(p.Item.x == 3)
		check(p.Item.y == 2)
		check(p.get().x == 3)
		check(p.get().y == 2)
	}
}

//
// Lambdas
//

func iterateOneToTen(f func(int)) {
	for i := 1; i <= 10; i++ {
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
		for i := 0; i < len(arr); i++ {
			sum += arr[i]
		}
		check(sum == 10)
		check(arr[1] == 2)
		setSecondElementToThree(&arr)
		check(arr[1] == 3)
	}
	{
		stuff := [...]int{1, 2, 3}
		check(len(stuff) == 3)
		sum := 0
		for i, elem := range stuff {
			check(i+1 == elem)
			sum += elem
		}
		check(sum == 6)
		// Other cases of for-range are checked in `testSlices`
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
		h := HasArray{}
		check(len(h.arr) == 4)
		check(h.arr[0] == 0)
		check(h.arr[1] == 0)
		check(h.arr[2] == 0)
		check(h.arr[3] == 0)
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
		s := [][]int{{1}, {}, {3, 4}}
		check(len(s) == 3)
		check(len(s[0]) == 1)
		check(s[0][0] == 1)
		check(len(s[1]) == 0)
		check(len(s[2]) == 2)
		check(s[2][0] == 3)
		check(s[2][1] == 4)
	}
	{
		stuff := []int{1, 2}
		stuff = append(stuff, 3)
		check(len(stuff) == 3)
		{
			sum := 0
			for i, elem := range stuff {
				check(i+1 == elem)
				sum += elem
			}
			check(sum == 6)
		}
		{
			sum := 0
			for i := range stuff {
				sum += i
			}
			check(sum == 3)
		}
		{
			sum := 0
			for _, elem := range stuff {
				sum += elem
			}
			check(sum == 6)
		}
		{
			count := 0
			for range stuff {
				count += 1
			}
			check(count == 3)
		}
		{
			stuff = []int{}
			count := 0
			for range stuff {
				count += 1
			}
			check(count == 0)
		}
	}
}

//
// Seq (generic slice with own methods)
//

type Seq[T any] []T

func (s *Seq[T]) len() int {
	return len(*s)
}

func (s *Seq[T]) add(val T) {
	*s = append(*s, val)
}

type Increr[T any] interface {
	*T
	incr()
}

func incrSeq[T any, PT Increr[T]](s *Seq[T]) {
	for i := range *s {
		PT(&(*s)[i]).incr()
	}
}

type SingleIncr struct {
	val int
}

func (s *SingleIncr) incr() {
	s.val += 1
}

type DoubleIncr struct {
	val int
}

func (s *DoubleIncr) incr() {
	s.val += 2
}

func testSeqs() {
	{
		s := Seq[int]{}
		check(s.len() == 0)
		s.add(1)
		s.add(2)
		check(s.len() == 2)
		check(s[0] == 1)
		check(s[1] == 2)
	}
	{
		s := Seq[int]{1, 2, 3}
		check(s.len() == 3)
		sum := 0
		for i, elem := range s {
			check(i+1 == elem)
			sum += elem
		}
		check(sum == 6)
	}
	{
		s := Seq[Point]{{1, 2}, {3, 4}}
		check(s.len() == 2)
		check(s[0].x == 1)
		check(s[0].y == 2)
		check(s[1].x == 3)
		check(s[1].y == 4)
		s.add(Point{5, 6})
		check(s[2].x == 5)
		check(s[2].y == 6)
	}
	{
		s := Seq[Point]{{x: 1, y: 2}, {x: 3, y: 4}}
		check(s.len() == 2)
	}
	{
		s := Seq[Seq[int]]{{1}, {}, {3, 4}}
		check(s.len() == 3)
		check(len(s[0]) == 1)
		check(s[0][0] == 1)
		check(s[1].len() == 0)
		check(s[2].len() == 2)
		check(s[2][0] == 3)
		check(s[2][1] == 4)
	}
	{
		s := Seq[SingleIncr]{{1}, {2}, {3}}
		incrSeq(&s)
		check(s[0].val == 2)
		check(s[1].val == 3)
		check(s[2].val == 4)
	}
	{
		s := Seq[DoubleIncr]{{1}, {2}, {3}}
		incrSeq(&s)
		check(s[0].val == 3)
		check(s[1].val == 4)
		check(s[2].val == 5)
	}
}

//
// Global variables
//

var globalY = globalX - three()
var globalX, globalZ = initialGlobalX, 14
var globalSlice []int

const initialGlobalX = 23

func setGlobalXToFortyTwo() {
	globalX = 42
}

func checkGlobalXIsFortyTwo() {
	check(globalX == 42)
}

func three() int {
	return 3
}

func isGlobalSliceEmpty() bool {
	return len(globalSlice) == 0
}

func apply(val int, fn func(int) int) int {
	return fn(val)
}

var globalApplied = apply(3, func(i int) int { return 2 * i })

type Enum int

const (
	ZeroEnum Enum = 0
	OneEnum       = 1
	TwoEnum       = 2
)

func testGlobalVariables() {
	{
		check(globalX == 23)
		check(globalY == 20)
		check(globalZ == 14)
		setGlobalXToFortyTwo()
		checkGlobalXIsFortyTwo()
		check(initialGlobalX == 23)
	}
	{
		check(isGlobalSliceEmpty())
		globalSlice = append(globalSlice, 1)
		globalSlice = append(globalSlice, 2)
		check(len(globalSlice) == 2)
		check(globalSlice[0] == 1)
		check(globalSlice[1] == 2)
		check(!isGlobalSliceEmpty())
	}
	{
		check(globalApplied == 6)
	}
	{
		check(ZeroEnum == 0)
		check(OneEnum == 1)
		check(TwoEnum == 2)
	}
}

//
// Imports
//

func testImports() {
	{
		f := foo.Foo{}
		check(f.Val() == 0)
	}
	{
		f := foo.NewFoo(42)
		check(f.Val() == 42)
	}
	{
		b := foo.Bar{X: 2, Y: 3}
		check(b.X == 2)
		check(b.Y == 3)
	}
}

//
// Externs
//

//gx:extern rect::NUM_VERTICES
const RectNumVertices = 0 // Ensure use of actual C++ constant value

//gx:extern rect::Rect
type Rect struct {
	X, Y          float32
	Width, Height float32
}

//gx:extern rect::area
func area(r Rect) float32

//gx:extern rect::area
func (r Rect) area() float32

func testExterns() {
	{
		check(RectNumVertices == 4)
		r := Rect{X: 100, Y: 100, Width: 20, Height: 30}
		check(r.X == 100)
		check(r.Y == 100)
		check(r.Width == 20)
		check(r.Height == 30)
		check(area(r) == 600)
		check(r.area() == 600)
	}
	{
		check(person.Population == 0)
		p := person.NewPerson(20, 100)
		check(person.Population == 1)
		check(p.Age() == 20)
		check(p.Health() == 100)
		p.Grow()
		check(p.Age() == 21)
		check(p.GXValue == 42)
		check(p.GetAgeAdder()(1) == 22)
	}
}

//
// Conversions
//

func testConversions() {
	{
		f := float32(2.2)
		i := int(f)
		check(i == 2)
		d := 2.2
		check(f-float32(d) == 0)
	}
	{
		slice := []int{1, 2}
		seq := Seq[int](slice)
		seq.add(3)
		check(seq.len() == 3)
		check(seq[0] == 1)
		check(seq[1] == 2)
		check(seq[2] == 3)
	}
}

//
// Meta
//

type Nums struct {
	A, B, C int
	D       int `attribs:"twice"`
}

//gx:extern sumFields
func sumFields(val interface{}) int

func testMeta() {
	n := Nums{1, 2, 3, 4}
	check(sumFields(n) == 14)
}

//
// Defaults
//

type HasDefaults struct {
	foo   int     `default:"42"`
	bar   float32 `default:"6.4"`
	point Point   `default:"{ 1, 2 }"`
}

func testDefaults() {
	h := HasDefaults{}
	check(h.foo == 42)
	check(h.bar == 6.4)
	check(h.point.x == 1)
	check(h.point.y == 2)
}

//
// Strings
//

//gx:extern std::strcmp
func strcmp(a, b string) int

type HasString struct {
	s string
}

func testStrings() {
	{
		s0 := ""
		check(len(s0) == 0)
		check(strcmp(s0, "") == 0)

		s1 := "foo"
		check(len(s1) == 3)
		check(s1[0] == 'f')
		check(s1[1] == 'o')
		check(s1[2] == 'o')
		check(strcmp(s1, "foo") == 0)

		s2 := "foo"
		check(strcmp(s1, s2) == 0)
		check(strcmp(s1, "nope") != 0)
		check(strcmp(s1, "foo") == 0)
		check(s1 == s2)
		check(s1 != "nope")
		check(s1 == string("foo"))
		check(s1 != string("fao"))

		s3 := s2
		check(strcmp(s1, s3) == 0)

		sum := 0
		for i, c := range s3 {
			sum += i
			if i == 0 {
				check(c == 'f')
			}
			if i == 1 {
				check(c == 'o')
			}
			if i == 2 {
				check(c == 'o')
			}
		}
		check(sum == 3)
	}
	{
		h0 := HasString{}
		check(len(h0.s) == 0)
		check(strcmp(h0.s, "") == 0)

		h1 := HasString{"foo"}
		check(len(h1.s) == 3)
		check(h1.s[0] == 'f')
		check(h1.s[1] == 'o')
		check(h1.s[2] == 'o')
		check(strcmp(h1.s, "foo") == 0)

		h2 := HasString{"foo"}
		check(strcmp(h1.s, h2.s) == 0)
		check(strcmp(h1.s, HasString{"nope"}.s) != 0)
		check(strcmp(h1.s, HasString{"foo"}.s) == 0)

		h3 := h2
		check(strcmp(h1.s, h3.s) == 0)
	}
}

//
// Defer
//

func testDefer() {
	x := 0
	{
		defer func() { x = 1 }()
		check(x == 0)
	}
	check(x == 1)

	setXTo2 := func() {
		x = 2
	}
	{
		defer setXTo2()
		check(x == 1)
	}
	check(x == 2)

	y := 0
	setYTo1AndReturn5 := func() int {
		y = 1
		return 5
	}
	setXToValue := func(val int) {
		x = val
	}
	{
		defer setXToValue(setYTo1AndReturn5())
		check(y == 0)
	}
	check(y == 1)
	check(x == 5)
}

//
// Main
//

func main() {
	testFib()
	testUnary()
	testVariables()
	testIncDec()
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
	testGlobalVariables()
	testImports()
	testExterns()
	testConversions()
	testMeta()
	testDefaults()
	testStrings()
	testDefer()
}
