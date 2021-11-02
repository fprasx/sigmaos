package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	db "ulambda/debug"
	"ulambda/fslib"
	np "ulambda/ninep"
	"ulambda/proc"
	"ulambda/procinit"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %v sleep_length out <native>\n", os.Args[0])
		os.Exit(1)
	}
	l, err := MakeSleeper(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: error %v", os.Args[0], err)
		os.Exit(1)
	}
	l.Work()
}

type Sleeper struct {
	*fslib.FsLib
	proc.ProcClnt
	native      bool
	sleepLength time.Duration
	output      string
}

func MakeSleeper(args []string) (*Sleeper, error) {
	if len(args) < 2 {
		return nil, errors.New("MakeSleeper: too few arguments")
	}
	s := &Sleeper{}
	db.Name("sleeper")
	s.FsLib = fslib.MakeFsLib("sleeper")
	s.ProcClnt = procinit.MakeProcClnt(s.FsLib, procinit.GetProcLayersMap())
	s.output = args[1]
	d, err := time.ParseDuration(args[0])
	if err != nil {
		log.Fatalf("Error parsing duration: %v", err)
	}
	s.sleepLength = d

	s.native = len(args) == 3 && args[2] == "native"

	db.DLPrintf("SCHEDL", "MakeSleeper: %v\n", args)

	if !s.native {
		err := s.Started(procinit.GetPid())
		if err != nil {
			log.Fatalf("Started: error %v\n", err)
		}
	}
	return s, nil
}

func (s *Sleeper) waitEvict() {
	if !s.native {
		err := s.WaitEvict(procinit.GetPid())
		if err != nil {
			log.Fatalf("Error WaitEvict: %v", err)
		}
		s.Exited(procinit.GetPid(), "EVICTED")
		os.Exit(0)
	}
}

func (s *Sleeper) Work() {
	go s.waitEvict()
	time.Sleep(s.sleepLength)
	err := s.MakeFile(s.output, 0777, np.OWRITE, []byte("hello"))
	if err != nil {
		log.Printf("Error: Makefile %v in Sleeper.Work: %v\n", s.output, err)
	}
	if !s.native {
		s.Exited(procinit.GetPid(), "OK")
	}
}
