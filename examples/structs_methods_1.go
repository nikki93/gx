package main

import (
	"fmt"
)

type Vec2 struct {
	x, y float32
}

type Entity int

type Position struct {
	Vec2
}

func (pos *Position) Add(ent Entity) {
	pos.x = 3
}

func main() {
	fmt.Println("Hello, world.")
}
