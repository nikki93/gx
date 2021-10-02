package foo

type Foo struct {
	val int
}

type Bar struct {
	X, Y int
}

func (f *Foo) Val() int {
	return f.val
}

func NewFoo(val int) Foo {
	return Foo{val}
}
