int sumFields(auto &val) {
  auto sum = 0;
  forEachField(val, [&](auto fieldTag, auto &fieldVal) {
    sum += fieldVal;
  });
  return sum;
}
