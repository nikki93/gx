//gx:include "person/person.hh"
//gx:externs person::

package person

type Person struct {
	age    int
	health float32
}

func NewPerson(age int, health float32) Person

func (p Person) Age() float32

func (p Person) Health() int

func (p *Person) Grow()
