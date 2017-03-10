#!/bin/bash

mkdir -p bin/

go get -v

for target in windows:amd64 linux:amd64 darwin:amd64; do
  echo "Compiling $target"
  export GOOS=$(echo $target | cut -d: -f1) GOARCH=$(echo $target | cut -d: -f2)
  OUT=bin/$(basename $(echo $PWD))_${GOOS}_${GOARCH}
  if [ $GOOS == "windows" ]
  then
    OUT="$OUT.exe"
  fi
  bash -c "go build -ldflags '-w' -v -o $OUT ."
done
