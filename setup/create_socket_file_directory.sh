#!/bin/bash

# ソケットファイルを格納するためのディレクトリを作成するスクリプト

dir="/tmp/mecm2m"

if [ ! -d "$dir" ]; then
    mkdir -p "$dir"
fi

dir2="/tmp/mecm2m/link-process"

if [ ! -d "$dir2" ]; then
    mkdir -p "$dir2"
fi