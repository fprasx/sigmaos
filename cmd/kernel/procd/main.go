package main

import (
	"fmt"
	"os"

	"ulambda/linuxsched"
	"ulambda/procd"
)

//
// Requires Unix path to parent of "bin" directory (e.g., ulambda) so
// procd knows where to find its executables.  In the longer run this
// should probably be a lambda pathname, but since procd uses
// cmd.Run, which requires a Unix path, we use Unix pathnames.
//

func main() {
	linuxsched.ScanTopology()
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %v parent-of-bin <pprof-output-path> <util-path>\n", os.Args[0])
		os.Exit(1)
	}
	pprofPath := ""
	if len(os.Args) >= 3 {
		pprofPath = os.Args[2]
	}
	utilPath := ""
	if len(os.Args) >= 4 {
		utilPath = os.Args[3]
	}
	procd.RunProcd(os.Args[1], pprofPath, utilPath)
}
