#!/bin/bash
mkdir -p build
cd midiforward && go build -o ../build/midiforward ./cmd/main.go
