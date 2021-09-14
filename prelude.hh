#include <cstdio>


//
// Print
//

inline void print(int val) {
  std::printf("%d", val);
}

inline void print(bool val) {
  std::printf(val ? "true": "false");
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
