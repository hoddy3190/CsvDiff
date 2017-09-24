#!/bin/bash

rm -rf pkg
mkdir -p pkg/linux_amd64
mkdir -p pkg/darwin_amd64
mkdir -p pkg/windows_386

GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o pkg/linux_amd64/CsvDiff
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o pkg/darwin_amd64/CsvDiff
GOOS=windows GOARCH=386 go build -ldflags "-s -w" -o pkg/windows_386/CsvDiff.exe
