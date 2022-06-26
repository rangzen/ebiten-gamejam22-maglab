#!/usr/bin/env sh

# Source: https://github.com/jakecoffman/magnets

# abort on errors
set -e

rm -rf dist
mkdir -p dist
GOOS=js GOARCH=wasm go build -o dist/maglab.wasm *.go
cp wasm_exec.js index.html dist

cd dist
git init
git add -A
git commit -m 'deploy'
git push -f git@github.com:rangzen/ebiten-gamejam22-maglab.git main:gh-pages

cd -

# Create the zip file for the distribution to itch.io
zip dist.zip dist/*