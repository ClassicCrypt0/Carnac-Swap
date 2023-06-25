#!/bin/bash

# Set GOOS and GOARCH for Windows build
export GOOS=windows
export GOARCH=amd64

# Build binaries from monitor_src
for file in monitor_src/*.go
do
  go build -o $(basename "$file" .go).exe $file
done

# Build binaries from swap_src
#for file in swap_src/*.go
#do
  #go build -o $(basename "$file" .go).exe $file
#done

# Unset GOOS and GOARCH for non-Windows build
unset GOOS
unset GOARCH

# Build binaries from monitor_src
for file in monitor_src/*.go
do
  go build -o $(basename "$file" .go) $file
done

# Build binaries from swap_src
#for file in swap_src/*.go
#do
  #go build -o $(basename "$file" .go) $file
#done