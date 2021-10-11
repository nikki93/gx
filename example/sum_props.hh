void forEachProp(auto &val, auto &&func) {
}

int sumProps(auto &val) {
  auto sum = 0;
  forEachProp(val, [&](auto propTag, auto &propVal) {
    sum += propVal;
  });
  return sum;
}
