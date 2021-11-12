#include <cstdio>
#include <vector>


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
  (print(args), ...);
}

template<typename... Args>
void println(Args &&...args) {
  print(args...);
  print("\n");
}


//
// Array
//

template<typename T, int N>
struct Array {
  T data[N] {};

  T &operator[](int i) {
    return data[i];
  }

  T *begin() {
    return &data[0];
  }

  T *end() {
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
  std::vector<T> data;

  Slice() = default;

  Slice(std::initializer_list<T> l)
      : data(l) {
  }

  T &operator[](int i) {
    return data[i];
  }

  auto begin() {
    return data.begin();
  }

  auto end() {
    return data.end();
  }
};

template<typename T>
int len(const Slice<T> &s) {
  return s.data.size();
}

template<typename T, typename U>
Slice<T> &append(Slice<T> &s, U &&val) {
  s.data.push_back(val);
  return s;
}


//
// String
//

struct String {
  std::vector<char> data;

  String() {
    data.resize(1);
    data[0] = '\0';
  }

  String(const char *s) {
    auto len = std::strlen(s);
    data.resize(len + 1);
    std::memcpy(data.data(), s, len);
    data[len] = '\0';
  }

  operator const char *() const {
    return (const char *)data.data();
  }

  char &operator[](int i) {
    return data[i];
  }

  auto begin() {
    return data.begin();
  }

  auto end() {
    return data.end() - 1;
  }
};

inline int len(const String &s) {
  return s.data.size() - 1;
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
