package procinit

import (
	"log"
	"os"
	"runtime/debug"
	"strings"

	"ulambda/procbase"
	"ulambda/procdep"
	"ulambda/fslib"
	"ulambda/procidem"
	"ulambda/proc"
)

const (
	PROC_LAYERS = "PROC_LAYERS" // Environment variable in which to store layer configuration
)

const ( // Possible stackable layers. PROCBASE is always included by default
	PROCBASE = "PROCBASE"
	PROCIDEM = "PROCIDEM"
	PROCDEP  = "PROCDEP"
)

// Get proc layers from environment variables.
func GetProcLayersMap() map[string]bool {
	s := os.Getenv(PROC_LAYERS)
	// XXX Remove eventually, just here to make sure we don't miss anything
	if len(s) == 0 {
		debug.PrintStack()
		log.Fatalf("Error! Length 0 sched layers!")
	}
	ls := strings.Split(s, ",")
	layers := make(map[string]bool)
	for _, l := range ls {
		layers[l] = true
	}
	layers[PROCBASE] = true
	return layers
}

func GetProcLayersString() string {
	s := os.Getenv(PROC_LAYERS)
	// XXX Remove eventually, just here to make sure we don't miss anything
	if len(s) == 0 {
		debug.PrintStack()
		log.Fatalf("Error! Length 0 sched layers!")
	}
	return PROC_LAYERS + "=" + s
}

func SetProcLayers(layers map[string]bool) {
	os.Setenv(PROC_LAYERS, makeProcLayersString(layers))
}

func makeProcLayersString(layers map[string]bool) string {
	s := ""
	for l, _ := range layers {
		s += l
		s += ","
	}
	s = s[:len(s)-1]
	return s
}

// Make a generic ProcClnt with the desired layers.
func MakeProcClnt(fsl *fslib.FsLib, layers map[string]bool) proc.ProcClnt {
	var clnt proc.ProcClnt
	clnt = procbase.MakeProcBaseClnt(fsl)
	if _, ok := layers[PROCIDEM]; ok {
		clnt = procidem.MakeProcIdemClnt(fsl, clnt)
	}
	if _, ok := layers[PROCDEP]; ok {
		clnt = procdep.MakeProcDepClnt(fsl, clnt)
	}
	return clnt
}
