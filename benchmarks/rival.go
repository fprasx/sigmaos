package benchmarks

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	db "ulambda/debug"
	"ulambda/fslib"
	"ulambda/proc"
	"ulambda/procclnt"
)

type Rival struct {
	spawnsPerSec       int
	secs               float64
	sleepIntervalUsecs int
	killed             bool
	ninep              bool
	dim                string
	its                string
	*fslib.FsLib
	*procclnt.ProcClnt
}

func MakeRival(args []string) (*Rival, error) {
	if len(args) < 5 {
		return nil, errors.New("MakeRival: too few arguments")
	}
	log.Printf("MakeRival: %v\n", args)

	r := &Rival{}
	r.FsLib = fslib.MakeFsLib("rival")
	r.ProcClnt = procclnt.MakeProcClnt(r.FsLib)

	sps, err := strconv.Atoi(args[0])
	r.spawnsPerSec = sps
	if err != nil {
		db.DFatalf("Invalid num spawns per sec: %v, %v\n", args[0], err)
	}

	secs, err := strconv.Atoi(args[1])
	r.secs = float64(secs)
	if err != nil {
		db.DFatalf("Invalid num seconds: %v, %v\n", args[0], err)
	}

	r.sleepIntervalUsecs = 1000000 / r.spawnsPerSec
	if r.secs >= 0 {
		log.Printf("Spawning every %v usec(s) for %v secs", r.sleepIntervalUsecs, r.secs)
	} else {
		log.Printf("Spawning every %v usec(s) indefinitely", r.sleepIntervalUsecs)
	}

	if args[2] == "native" {
		r.ninep = false
	} else if args[2] == "ninep" {
		r.ninep = true
	} else {
		db.DFatalf("Unexpected rival spawn type: %v", args[2])
	}

	r.dim = args[3]
	if err != nil {
		db.DFatalf("Invalid dimension: %v, %v\n", args[3], err)
	}

	r.its = args[4]
	if err != nil {
		db.DFatalf("Invalid iterations: %v, %v\n", args[4], err)
	}

	return r, nil
}

func (r *Rival) spawnSpinner(pid proc.Tpid) {
	if r.ninep {
		a := proc.MakeProcPid(pid, "user/c-spinner", []string{r.dim, r.its})
		start := time.Now()
		err := r.Spawn(a)
		if err != nil {
			db.DFatalf("couldn't spawn ninep spinner %v: %v\n", pid, err)
		}
		go func() {
			_, err := r.WaitExit(pid)
			if err != nil {
				log.Printf("Error running lambda: %v", err)
			}
			end := time.Now()
			elapsed := end.Sub(start)
			log.Printf("Ninep elapsed time: %f usec(s)\n", float64(elapsed.Microseconds()))
		}()
	} else {
		cmd := exec.Command(".user/c-spinner", []string{pid.String(), r.dim, r.its, "native"}...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		start := time.Now()
		err := cmd.Start()
		if err != nil {
			log.Printf("Error starting native spinner: %v, %v\n", pid, err)
		}
		go func() {
			err := cmd.Wait()
			if err != nil {
				log.Printf("Error running command: %v", err)
			}
			end := time.Now()
			elapsed := end.Sub(start)
			log.Printf("Ninep elapsed time: %f usec(s)\n", float64(elapsed.Microseconds()))
		}()
	}
}

func (r *Rival) Work() {
	//	m := linuxsched.CreateCPUMaskOfOne(0)
	//	linuxsched.SchedSetAffinityAllTasks(os.Getpid(), m)
	start := time.Now()
	for {
		// Check if we're done
		if r.secs >= 0 && time.Now().Sub(start).Seconds() >= r.secs {
			break
		}
		pid := proc.GenPid()
		r.spawnSpinner(pid)
		time.Sleep(time.Duration(r.sleepIntervalUsecs) * time.Microsecond)
	}
	r.killed = true
}
