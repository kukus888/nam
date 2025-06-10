#!/bin/bash

CGO_ENABLED=0 go build -ldflags "-extldflags=-static" -o nam.static.linux64.bin
CGO_ENABLED=0 go build -o nam.dynamic.linux64.bin