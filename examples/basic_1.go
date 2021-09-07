package main

import "fmt"

type Foo struct {
	X int
	Y [32]byte
}

func main() {
	x := 42
	println("hello, world!")
	println(x)

	fmt.Println("hello, world!")
}
