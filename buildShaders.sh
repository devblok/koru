#!/usr/bin/env bash

BINARY_FOLDER=$1

for shader in $(ls shaders)
do
    glslangValidator -V shaders/$shader -o $BINARY_FOLDER/$shader.spv -e main --source-entrypoint main
done