package main

import (
	"os"

	"sigmaos/cachesrv"
	db "sigmaos/debug"
)

func main() {
	if len(os.Args) < 3 {
		db.DFatalf("Usage: %v cachedir public [shrdpn]", os.Args[0])
	}
	if err := cachesrv.RunCacheSrv(os.Args); err != nil {
		db.DFatalf("Start %v err %v\n", os.Args[0], err)
	}
}
