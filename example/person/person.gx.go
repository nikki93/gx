//gx:include "person/person.hh"
//gx:externs person::

package person

type Person struct {
	age    int
	health float32
}

func NewPerson(age int, health float32) Person

//gx:extern person::GetAge
func (p Person) Age() float32

//gx:extern person::GetHealth
func (p Person) Health() int

func (p *Person) Grow()
