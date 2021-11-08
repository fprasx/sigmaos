package main

import (
	"log"

	"ulambda/fslibsrv"
	"ulambda/linuxsched"
	"ulambda/named"
	"ulambda/proc"
	"ulambda/procinit"
	"ulambda/seccomp"
)

func main() {
	linuxsched.ScanTopology()
	// started as a ulambda
	name := named.MEMFS + "/" + proc.GetPid()
	mfs, err := fslibsrv.StartMemFs(name, name)
	if err != nil {
		log.Fatalf("StartMemFs %v\n", err)
	}
	sclnt := procinit.MakeProcClnt(mfs.FsLib, procinit.GetProcLayersMap())
	sclnt.Started(proc.GetPid())
	seccomp.LoadFilter()
	mfs.Wait()
	sclnt.Exited(proc.GetPid(), "OK")
}
