#pragma once

#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <new>
#include <utility>


namespace gx {


//
// Print
//

inline void print(bool val) {
  std::printf(val ? "true" : "false");
}

inline void print(int val) {
  std::printf("%d", val);
}

inline void print(float val) {
  std::printf("%g", val);
}

inline void print(double val) {
  std::printf("%f", val);
}

inline void print(const char *val) {
  std::printf("%s", val);
}

template<typename A, typename B, typename... Args>
void print(A &a, B &&b, Args &&...args) {
  print(a);
  print(b);
  (print(std::forward<Args>(args)), ...);
}

template<typename... Args>
void println(Args &&...args) {
  print(std::forward<Args>(args)...);
  print("\n");
}

template<typename... Args>
void fatal(Args &&...args) {
  println(std::forward<Args>(args)...);
  std::fflush(stdout);
  std::abort();
}


//
// Pointer
//

template<typename T>
T &deref(T *ptr) {
#ifndef GX_NO_CHECKS
  if (!ptr) {
    fatal("gx: nil pointer dereference");
  }
#endif
  return *ptr;
}

template<typename T>
const T &deref(const T *ptr) {
  return deref(const_cast<T *>(ptr));
}


//
// Array
//

template<typename T, int N>
struct Array {
  T data[N] {};

  T &operator[](int i) {
#ifndef GX_NO_CHECKS
    if (!(0 <= i && i < N)) {
      fatal("gx: array index out of bounds");
    }
#endif
    return data[i];
  }

  const T &operator[](int i) const {
    return const_cast<Array &>(*this)[i];
  }

  T *begin() {
    return &data[0];
  }

  const T *begin() const {
    return &data[0];
  }

  T *end() {
    return &data[N];
  }

  const T *end() const {
    return &data[N];
  }
};

template<typename T, int N>
constexpr int len(const Array<T, N> &a) {
  return N;
}


//
// Slice
//

template<typename T>
struct Slice {
  T *data = nullptr;
  int size = 0;
  int capacity = 0;

  Slice() = default;

  Slice(const Slice &other) {
    if (this != &other) {
      copyFrom(other.data, other.size);
    }
  }

  Slice &operator=(const Slice &other) {
    if (this != &other) {
      destruct();
      copyFrom(other.data, other.size);
    }
    return *this;
  }

  Slice(Slice &&other) {
    if (this != &other) {
      moveFrom(other);
    }
  }

  Slice &operator=(Slice &&other) {
    if (this != &other) {
      destruct();
      moveFrom(other);
    }
    return *this;
  }

  Slice(std::initializer_list<T> l) {
    copyFrom(l.begin(), l.size());
  }

  ~Slice() {
    destruct();
  }

  void copyFrom(const T *data_, int size_) {
    data = (T *)std::malloc(sizeof(T) * size_);
    size = size_;
    capacity = size_;
    for (auto i = 0; auto &elem : *this) {
      new (&elem) T(data_[i++]);
    }
  }

  void moveFrom(Slice &other) {
    data = other.data;
    size = other.size;
    capacity = other.capacity;
    other.data = nullptr;
    other.size = 0;
    other.capacity = 0;
  }

  void destruct() {
    for (auto &elem : *this) {
      elem.~T();
    }
    std::free(data);
  }

  T &operator[](int i) {
#ifndef GX_NO_CHECKS
    if (!(0 <= i && i < size)) {
      fatal("gx: slice index out of bounds");
    }
#endif
    return data[i];
  }

  const T &operator[](int i) const {
    return const_cast<Slice &>(*this)[i];
  }

  T *begin() {
    return &data[0];
  }

  const T *begin() const {
    return &data[0];
  }

  T *end() {
    return &data[size];
  }

  const T *end() const {
    return &data[size];
  }
};

template<typename T>
int len(const Slice<T> &s) {
  return s.size;
}

template<typename T>
void insert(Slice<T> &s, int i, T val) {
#ifndef GX_NO_CHECKS
  if (!(0 <= i && i <= s.size)) {
    fatal("gx: slice index out of bounds");
  }
#endif
  auto moveCount = s.size - i;
  ++s.size;
  if (s.size > s.capacity) {
    s.capacity = s.capacity == 0 ? 2 : s.capacity << 1;
    s.data = (T *)std::realloc(s.data, sizeof(T) * s.capacity);
  }
  std::memmove(&s.data[i + 1], &s.data[i], sizeof(T) * moveCount);
  new (&s.data[i]) T(std::move(val));
}

template<typename T>
Slice<T> &append(Slice<T> &s, T val) {
  insert(s, s.size, std::move(val));
  return s;
}

template<typename T>
T &append(Slice<T> &s) {
  insert(s, s.size, T {});
  return s[len(s) - 1];
}

template<typename T>
void remove(Slice<T> &s, int i) {
#ifndef GX_NO_CHECKS
  if (!(0 <= i && i < s.size)) {
    fatal("gx: slice index out of bounds");
  }
#endif
  auto moveCount = s.size - (i + 1);
  s.data[i].~T();
  std::memmove(&s.data[i], &s.data[i + 1], sizeof(T) * moveCount);
  --s.size;
}


//
// String
//

struct String {
  Slice<char> slice;

  String()
      : slice({ '\0' }) {
  }

  String(const char *s) {
    slice.copyFrom(s, std::strlen(s) + 1);
  }

  operator const char *() const {
    return (const char *)slice.data;
  }

  char &operator[](int i) {
#ifndef GX_NO_CHECKS
    if (!(0 <= i && i < slice.size)) {
      fatal("gx: string index out of bounds");
    }
#endif
    return slice[i];
  }

  auto begin() {
    return slice.begin();
  }

  auto end() {
    return slice.end() - 1;
  }
};

inline int len(const String &s) {
  return s.slice.size - 1;
}

inline bool operator==(const String &a, const String &b) {
  auto aSize = a.slice.size;
  if (aSize != b.slice.size) {
    return false;
  }
  return !std::memcmp(a.slice.data, b.slice.data, aSize);
}

inline bool operator==(const String &a, const char *b) {
  return !std::strcmp(a, b);
}


//
// Meta
//

#ifndef GX_FIELD_ATTRIBS
struct FieldAttribs {
  const char *name;
};
#else
using FieldAttribs = GX_FIELD_ATTRIBS;
#endif

template<typename T, int N>
struct FieldTag {};


}
