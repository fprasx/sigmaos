#!/bin/bash

usage() {
  echo "Usage: $0 [--norace] [--vet] [--parallel] [--gopath GO] [--target TARGET] [--bins BINS] kernel|user" 1>&2
}

RACE="-race"
CMD="build"
TARGET="local"
GO="go"
PARALLEL=""
WHAT=""
BINS=""
while [[ "$#" -gt 0 ]]; do
  case "$1" in
  --norace)
    shift
    RACE=""
    ;;
  --vet)
    shift
    CMD="vet"
    ;;
  --gopath)
    shift
    GO="$1"
    shift
    ;;
  --target)
    shift
    TARGET="$1"
    shift
    ;;
  --parallel)
    shift
    PARALLEL="--parallel"
    ;;
  --bins)
    shift
    BINS="$1"
    shift
    ;;
  -help)
    usage
    exit 0
    ;;
  kernel|user|linux)
    WHAT=$1
    shift
    ;;
  *)
   echo "unexpected argument $1"
   usage
   exit 1
  esac
done

if [ $# -gt 0 ]; then
    usage
    exit 1
fi
echo $WHAT

if [[ $WHAT == "kernel" ]]; then
    mkdir -p bin/kernel
    mkdir -p bin/linux
    WHAT="kernel linux"
elif [[ $WHAT == "user" ]]; then
    mkdir -p bin/user
else
    mkdir -p bin/linux
    WHAT="linux"
fi

LDF="-X sigmaos/sigmap.Target=$TARGET"

for k in $WHAT; do
    echo "Building $k components"
    if [[ $WHAT == "user" ]] && [[ $BINS != "" ]]; then
        FILES=$BINS
    else
        FILES=`ls cmd/$k`
    fi
    
    for f in $FILES;  do
        # XXX delete when removing obselete code
        if [[ $f == "sigmamgr" ]] || [[ $f == "memfs-raft-replica" ]] ; then
            continue
        fi
        if [ $CMD == "vet" ]; then
            echo "$GO vet cmd/$k/$f/main.go"
            $GO vet cmd/$k/$f/main.go
        else
            build="$GO build -ldflags=\"$LDF\" $RACE -o bin/$k/$f cmd/$k/$f/main.go"
            echo $build
            if [ -z "$PARALLEL" ]; then
                eval "$build"
            else
                eval "$build" &
            fi
        fi
    done
done

wait

