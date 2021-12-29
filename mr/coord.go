package mr

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"ulambda/crash"
	db "ulambda/debug"
	"ulambda/fslib"
	np "ulambda/ninep"
	"ulambda/proc"
	"ulambda/procclnt"
	usync "ulambda/sync"
)

const (
	INPUTDIR = "name/s3/~ip/input/"
	MRDIR    = "name/mr"
	MDIR     = "name/mr/m"
	RDIR     = "name/mr/r"
	ROUT     = "name/mr/mr-out-"
	CLAIMED  = "-claimed"
	TIP      = "-tip"
	DONE     = "-done"

	// time interval (ms) for when a failure might happen. If too
	// frequent and they don't finish ever. XXX determine
	// dynamically
	CRASHMAPPER  = 10000
	CRASHREDUCER = 10000
	CRASHCOORD   = 20000
)

func InitCoordFS(fsl *fslib.FsLib, nreducetask int) {
	for _, n := range []string{MRDIR, MDIR, RDIR, MDIR + CLAIMED, RDIR + CLAIMED, MDIR + TIP, RDIR + TIP, MDIR + DONE, RDIR + DONE} {
		if err := fsl.Mkdir(n, 0777); err != nil {
			log.Fatalf("Mkdir %v\n", err)
		}
	}

	// input directories for reduce tasks
	for r := 0; r < nreducetask; r++ {
		n := RDIR + "/" + strconv.Itoa(r)
		if err := fsl.Mkdir(n, 0777); err != nil {
			log.Fatalf("Mkdir %v err %v\n", n, err)
		}
	}
}

type Coord struct {
	*fslib.FsLib
	*procclnt.ProcClnt
	crashCoord  string
	nreducetask int
	crash       string
	mapperbin   string
	reducerbin  string
	lease       *usync.LeasePath
}

func MakeCoord(args []string) (*Coord, error) {
	if len(args) != 5 {
		return nil, errors.New("MakeCoord: too few arguments")
	}
	w := &Coord{}
	w.FsLib = fslib.MakeFsLib("coord-" + proc.GetPid())

	n, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("MakeCoord: nreducetask %v isn't int", args[1])
	}

	w.nreducetask = n
	w.mapperbin = args[1]
	w.reducerbin = args[2]
	w.crash = args[3]
	w.crashCoord = args[4]

	w.ProcClnt = procclnt.MakeProcClnt(w.FsLib)

	w.Started(proc.GetPid())

	w.lease = usync.MakeLeasePath(w.FsLib, MRDIR+"/lease-coord", 0)

	if w.crashCoord == "YES" {
		crash.Crasher(w.FsLib, CRASHCOORD)
	}

	return w, nil
}

func (w *Coord) mapper(task string) string {
	input := INPUTDIR + task
	a := proc.MakeProc(w.mapperbin, []string{w.crash, strconv.Itoa(w.nreducetask), input})
	err := w.Spawn(a)
	if err != nil {
		return err.Error()
	}
	ok, err := w.WaitExit(a.Pid)
	if err != nil {
		return err.Error()
	}
	return ok
}

func (w *Coord) reducer(task string) string {
	in := RDIR + TIP + "/" + task
	out := ROUT + task
	a := proc.MakeProc(w.reducerbin, []string{w.crash, in, out})
	err := w.Spawn(a)
	if err != nil {
		return err.Error()
	}
	ok, err := w.WaitExit(a.Pid)
	if err != nil {
		return err.Error()
	}
	return ok
}

func (w *Coord) claimEntry(dir string, st *np.Stat) (string, error) {
	from := dir + "/" + st.Name
	if err := w.Rename(from, dir+TIP+"/"+st.Name); err != nil {
		if err.Error() == "EOF" { // all errors, except not found?
			return "", err
		}
		// another coord claimed the task
		return "", nil
	}
	return st.Name, nil
}

func (w *Coord) getTask(dir string) (string, error) {
	t := ""
	stopped, err := w.ProcessDir(dir, func(st *np.Stat) (bool, error) {
		t1, err := w.claimEntry(dir, st)
		if err != nil {
			return false, err
		}
		if t1 == "" {
			return false, nil
		}
		t = t1
		return true, nil

	})
	if err != nil {
		return "", err
	}
	if stopped {
		return t, nil
	}
	return "", nil
}

type Ttask struct {
	task string
	ok   string
}

func (w *Coord) startTasks(dir string, ch chan Ttask, f func(string) string) int {
	n := 0
	for {
		t, err := w.getTask(dir)
		if err != nil {
			log.Fatalf("getTask %v err %v\n", dir, err)
		}
		if t == "" {
			break
		}
		n += 1
		go func() {
			db.DPrintf("start task %v\n", t)
			ok := f(t)
			ch <- Ttask{t, ok}
		}()
	}
	return n
}

func (w *Coord) processResult(dir string, res Ttask) {
	if res.ok == "OK" {
		// mark task as done
		log.Printf("%v: task done %v\n", db.GetName(), res.task)
		s := dir + TIP + "/" + res.task
		d := dir + DONE + "/" + res.task
		err := w.Rename(s, d)
		if err != nil {
			// an earlier instance already succeeded
			log.Printf("%v: rename %v to %v err %v\n", db.GetName(), s, d, err)
		}
	} else {
		// task failed; make it runnable again
		to := dir + "/" + res.task
		db.DPrintf("task %v failed %v\n", res.task, res.ok)
		if err := w.Rename(dir+TIP+"/"+res.task, to); err != nil {
			log.Fatalf("%v: rename to %v err %v\n", db.GetName(), to, err)
		}
	}
}

func (w *Coord) stragglers(dir string, ch chan Ttask, f func(string) string) {
	sts, err := w.ReadDir(dir + TIP) // XXX handle one entry at the time?
	if err != nil {
		log.Fatalf("recover: ReadDir %v err %v\n", dir+TIP, err)
	}
	n := 0
	for _, st := range sts {
		n += 1
		go func() {
			log.Printf("%v: start straggler task %v\n", db.GetName(), st.Name)
			ok := f(st.Name)
			ch <- Ttask{st.Name, ok}
		}()
	}
}

func (w *Coord) recover(dir string) {
	sts, err := w.ReadDir(dir + TIP) // XXX handle one entry at the time?
	if err != nil {
		log.Fatalf("recover: ReadDir %v err %v\n", dir+TIP, err)
	}

	// just treat all tasks in progress as failed; too aggressive, but correct.
	for _, st := range sts {
		log.Printf("%v: recover %v\n", db.GetName(), st.Name)
		to := dir + "/" + st.Name
		if w.Rename(dir+TIP+"/"+st.Name, to) != nil {
			// an old, disconnected coord may do this too,
			// if one of its tasks fails
			log.Printf("%v: rename to %v err %v\n", db.GetName(), to, err)
		}
	}
}

func (w *Coord) phase(dir string, f func(string) string) {
	ch := make(chan Ttask)
	straggler := false
	for n := w.startTasks(dir, ch, f); n > 0; n-- {
		res := <-ch
		w.processResult(dir, res)
		if res.ok != "OK" {
			n += w.startTasks(dir, ch, f)
		}
		if n == 2 && !straggler { // XXX percentage of total computation
			straggler = true
			w.stragglers(dir, ch, f)
		}
	}
}

func (w *Coord) Work() {
	// Try to become the primary coordinator.  Backup coordinators
	// will be able to acquire the lease if the primary fails or
	// is partitioned.
	w.lease.WaitWLease([]byte{})
	defer w.lease.ReleaseWLease()

	log.Printf("%v: primary\n", db.GetName())

	w.recover(MDIR)
	w.recover(RDIR)

	w.phase(MDIR, w.mapper)
	log.Printf("%v: Reduce phase\n", db.GetName())
	w.phase(RDIR, w.reducer)

	w.Exited(proc.GetPid(), "OK")
}
