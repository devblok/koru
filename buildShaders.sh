#!/usr/bin/env bash

BINARY_FOLDER=$1

for shader in $(ls src/shaders)
do
    glslangValidator -s -V src/shaders/$shader -o $BINARY_FOLDER/$shader.spv -e main --source-entrypoint main
done