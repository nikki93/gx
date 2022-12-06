#!/bin/bash

set -e

PLATFORM="macOS"
CLANG="clang++"
GO="go"
TIME="time"
TIME_TOTAL="time"
EXE=""

if [[ -f /proc/version ]]; then
  if grep -q Linux /proc/version; then
    PLATFORM="lin"
    TIME="time --format=%es\n"
    TIME_TOTAL="time --format=total\t%es\n"
  fi
  if grep -q Microsoft /proc/version; then
    PLATFORM="win"
    CLANG="clang++.exe"
    GO="go.exe"
    EXE=".exe"
  fi
fi
CLANG="$TIME $CLANG"
GO="$TIME $GO"

case "$1" in
  # Desktop
  release)
    mkdir -p build
    rm -rf build/*

    $GO build gx.go

    $TIME ./gx$EXE ./example build/example
    if [[ -f build/example.gx.cc ]]; then
      $CLANG -std=c++20 -Wall -O3 -Iexample -o build/example build/example.*.cc
    fi
    $TIME ./gx$EXE ./example/gxsl build/example_gxsl
    if [[ -f build/example_gxsl.gx.cc ]]; then
      $CLANG -std=c++20 -Wall -O3 -o build/example_gxsl build/example_gxsl.*.cc
    fi

    if [[ -f build/example ]]; then
      ./build/example
    fi
    if [[ -f build/example_gxsl ]]; then
      ./build/example_gxsl > ./build/example_gxsl_output.frag
      glslangValidator$EXE ./build/example_gxsl_output.frag | sed "s/^ERROR: 0/.\/build\/example_gxsl_output.frag/g" | sed "/\.frag$/d"
    fi

    exit 1
    ;;
esac
