#!/bin/bash
GOOS=darwin GOARCH=arm64 go build -o eniacsim_arm64
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o eniacsim_amd64
lipo -create -output eniacsim eniacsim_amd64 eniacsim_arm64
rm eniacsim_amd64 eniacsim_arm64