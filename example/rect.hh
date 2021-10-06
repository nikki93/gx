#pragma once


namespace rect {

inline constexpr auto NUM_VERTICES = 4;

struct Rect {
  float x, y;
  float width, height;
};

inline float area(Rect r) {
  return r.width * r.height;
}

}
