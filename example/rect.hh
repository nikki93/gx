#pragma once


namespace rect {

struct Rect {
  float x, y;
  float width, height;
};

inline float area(Rect r) {
  return r.width * r.height;
}

}
