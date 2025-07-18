#!/bin/bash

# 创建构建目录
mkdir -p build

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o build/yaml-editor-windows-amd64.exe

# macOS (amd64)
GOOS=darwin GOARCH=amd64 go build -o build/yaml-editor-darwin-amd64

# macOS (arm64, for Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o build/yaml-editor-darwin-arm64

# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o build/yaml-editor-linux-amd64

echo "Build completed. Files in build/"