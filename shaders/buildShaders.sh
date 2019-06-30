#!/usr/bin/env bash

BINARY_FOLDER=../$1

for shader in $(ls)
do
    glslangValidator -V $shader -o $BINARY_FOLDER/$shader.spv 
done