//gx:include "person/person.hh"
//gx:externs person::

package person

type Person struct {
	age     int
	health  float32
	GXValue int //gx:extern cppValue
}

var Population int

func NewPerson(age int, health float32) Person

//gx:extern person::GetAge
func (p Person) Age() float32

//gx:extern person::GetHealth
func (p Person) Health() int

func (p *Person) Grow()

//gx:extern person::INVALID
type AgeAdder func(i int) int

func (p *Person) GetAgeAdder() AgeAdder
