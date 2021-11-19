struct SumFieldsAttribs {
  const char *name;
  bool twice = false;
};

#define GX_FIELD_ATTRIBS SumFieldsAttribs

int sumFields(auto &val) {
  auto sum = 0;
  forEachField(val, [&](auto fieldTag, auto &fieldVal) {
    if constexpr (fieldTag.attribs.twice) {
      sum += 2 * fieldVal;
    } else {
      sum += fieldVal;
    }
  });
  return sum;
}
