#pragma once


namespace person {

struct Person {
  int age;
  float health;
  int cppValue = 42;
};

inline int Population = 0;

inline Person NewPerson(int age, float health) {
  ++Population;
  return Person { age, health };
}

inline int GetAge(Person p) {
  return p.age;
}

inline float GetHealth(Person p) {
  return p.health;
}

inline void Grow(Person *p) {
  p->age++;
}

}
