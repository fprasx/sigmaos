package namedv1

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"

	"sigmaos/container"
	db "sigmaos/debug"
	"sigmaos/fslibsrv"
	"sigmaos/proc"
	"sigmaos/sesssrv"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
)

var (
	dialTimeout = 5 * time.Second

	endpoints = []string{"127.0.0.1:2379", "localhost:22379", "localhost:32379"}
)

var nd *Named

type Named struct {
	*sigmaclnt.SigmaClnt
	*sesssrv.SessSrv
	mu    sync.Mutex
	clnt  *clientv3.Client
	sess  *concurrency.Session
	job   string
	crash int
}

func Run(args []string) error {
	bootNamed := len(args) == 1
	db.DPrintf(db.NAMEDV1, "BootNamed %v\n", bootNamed)
	if !(len(args) == 1 || len(args) == 3) {
		return fmt.Errorf("%v: wrong number of arguments %v", args[0], args)
	}
	nd = &Named{}
	if !bootNamed {
		nd.job = args[1]
		crashing, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("%v: crash %v isn't int", args[0], args[2])
		}
		nd.crash = crashing
		sc, err := sigmaclnt.MkSigmaClnt(proc.GetPid().String())
		if err != nil {
			return err
		}
		nd.SigmaClnt = sc
		db.DPrintf(db.NAMEDV1, "start %d\n", len(args))
		nd.Started()
		db.DPrintf(db.NAMEDV1, "started %d\n", len(args))
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		db.DFatalf("Error clientv3 %v\n", err)
	}
	nd.clnt = cli
	s, err := concurrency.NewSession(cli)
	if err != nil {
		db.DFatalf("Error sess %v\n", err)
	}
	defer cli.Close()

	nd.sess = s

	ip, err := container.LocalIP()
	if err != nil {
		db.DFatalf("LocalIP %v %v\n", sp.UX, err)
	}

	go nd.waitExit()

	fn := "named-election"
	// fn := fmt.Sprintf("job-%s-election", nd.job))
	db.DPrintf(db.NAMEDV1, "candidate %v %v\n", proc.GetPid().String(), fn)

	electclnt := concurrency.NewElection(nd.sess, fn)

	if err := electclnt.Campaign(context.TODO(), proc.GetPid().String()); err != nil {
		db.DFatalf("Campaign err %v\n", err)
	}

	db.DPrintf(db.NAMEDV1, "leader %v\n", proc.GetPid().String())

	root := rootDir()

	srv, err := fslibsrv.MakeReplServer(root, ip+":0", sp.NAMEDV1, "namedv1", nil)
	if err != nil {
		db.DFatalf("Error MakeMemFs: %v", err)
	}
	nd.SessSrv = srv

	srv.Serve()

	db.DPrintf(db.NAMEDV1, "terminate\n")

	srv.Done()

	return nil
}

func (nd *Named) waitExit() {
	err := nd.WaitEvict(proc.GetPid())
	if err != nil {
		db.DFatalf("Error WaitEvict: %v", err)
	}
	db.DPrintf(db.NAMEDV1, "candidate %v evicted\n", proc.GetPid().String())
	nd.Exited(proc.MakeStatus(proc.StatusEvicted))
	os.Exit(0)
}
