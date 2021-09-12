#include <cstdio>


//
// Print
//

inline void print(int val) {
  std::printf("%d", val);
}

inline void print(const char *val) {
  std::printf("%s", val);
}

template<typename... Args>
void print(Args &&...args) {
  (print(args), ...);
}
