package schedd

import (
	//	"github.com/sasha-s/go-deadlock"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	db "ulambda/debug"
	"ulambda/fsclnt"
	"ulambda/fslib"
	np "ulambda/ninep"
	npo "ulambda/npobjsrv"
	"ulambda/npsrv"
)

// XXX monitor, boost

const (
	MAXLOAD      = 100 // XXX bogus, controls parallelism
	NO_OP_LAMBDA = "no-op-lambda"
)

type Sched struct {
	//	mu   deadlock.Mutex
	mu   sync.Mutex
	cond *sync.Cond
	load int // XXX bogus
	nid  uint64
	ls   map[string]*Lambda
	root *Obj
	done bool
	srv  *npsrv.NpServer
	*fslib.FsLib
}

func MakeSchedd() *Sched {
	sd := &Sched{}
	sd.cond = sync.NewCond(&sd.mu)
	sd.load = 0
	sd.nid = 0
	sd.ls = make(map[string]*Lambda)
	sd.root = sd.MakeObj([]string{}, np.DMDIR, nil).(*Obj)
	sd.root.time = time.Now().Unix()
	db.SetDebug(false)
	ip, err := fsclnt.LocalIP()
	if err != nil {
		log.Fatalf("LocalIP %v %v\n", fslib.SCHED, err)
	}
	sd.srv = npsrv.MakeNpServer(sd, ip+":0")
	fsl := fslib.MakeFsLib("sched")
	sd.FsLib = fsl
	err = fsl.PostService(sd.srv.MyAddr(), fslib.SCHED)
	if err != nil {
		log.Fatalf("PostService failed %v %v\n", fslib.SCHED, err)
	}
	return sd
}

func (sd *Sched) Connect(conn net.Conn) npsrv.NpAPI {
	return npo.MakeNpConn(sd, conn)
}

func (sd *Sched) Done() {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	sd.done = true
	sd.cond.Signal()
}

func (sd *Sched) Root() npo.NpObj {
	return sd.root
}

func (sd *Sched) Resolver() npo.Resolver {
	return nil
}

func (sd *Sched) String() string {
	s := ""
	for _, l := range sd.ls {
		l.mu.Lock()
		defer l.mu.Unlock()
		s += l.String()
		s += "\n"
	}
	return s
}

func (sd *Sched) uid() uint64 {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.nid += 1
	return sd.nid
}

func (sd *Sched) spawn(l *Lambda) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	sd.ls[l.Pid] = l
	sd.cond.Signal()
}

// Lock & find lambda
func (sd *Sched) findLambda(pid string) *Lambda {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.findLambdaS(pid)
}

func (sd *Sched) findLambdaS(pid string) *Lambda {
	l, _ := sd.ls[pid]
	return l
}

func (sd *Sched) delLambda(pid string) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	delete(sd.ls, pid)
}

func (sd *Sched) decLoad() {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	sd.load -= 1
	sd.cond.Signal()
}

// wakeup lambdas that have pid as an exit dependency
func (sd *Sched) wakeupExit(pid string) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	for _, l := range sd.ls {
		l.startExitDep(pid)
	}
	sd.cond.Signal()
}

// Caller holds sd lock
func (sd *Sched) findRunnable() *Lambda {
	for _, l := range sd.ls {
		if l.isRunnable() {
			return l
		}
	}
	return nil
}

// Caller holds sd lock
func (sd *Sched) findRunnableWaitingConsumer() *Lambda {
	for _, l := range sd.ls {
		if l.runnableWaitingConsumer() {
			return l
		}
	}
	return nil
}

// Select a random locald instance to run on
func (sd *Sched) selectLocaldIp() (string, error) {
	ips, err := sd.ReadDir(fslib.LOCALD_ROOT)
	if err != nil {
		log.Printf("Schedd error reading localds dir\n: %v", err)
		return "", err
	}
	n := rand.Int() % len(ips)
	return ips[n].Name, nil
}

func (sd *Sched) Scheduler() {
	sd.mu.Lock()
	for !sd.done {
		l := sd.findRunnableWaitingConsumer()
		if l != nil {
			// XXX don't count starting a consumer against load
			l.run()
			sd.load += 1
		} else {
			if sd.load < MAXLOAD {
				l = sd.findRunnable()
				if l != nil {
					l.run()
					sd.load += 1
				}
			}
			if l == nil || sd.load >= MAXLOAD {
				sd.cond.Wait()
			}
		}
	}
}
