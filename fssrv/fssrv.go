package fssrv

import (
	"fmt"
	"log"
	"runtime/debug"

	// db "ulambda/debug"
	db "ulambda/debug"
	"ulambda/fs"
	"ulambda/fslib"
	"ulambda/netsrv"
	np "ulambda/ninep"
	"ulambda/proc"
	"ulambda/procclnt"
	"ulambda/protsrv"
	"ulambda/repl"
	"ulambda/sesscond"
	"ulambda/session"
	"ulambda/stats"
	"ulambda/watch"
)

//
// There is one FsServer per memfsd. The FsServer has one ProtSrv per
// 9p channel (i.e., TCP connection); each channel has one or more
// sessions (one per client fslib on the same client machine).
//

type FsServer struct {
	addr  string
	root  fs.Dir
	mkps  protsrv.MkProtServer
	stats *stats.Stats
	st    *session.SessionTable
	sct   *sesscond.SessCondTable
	wt    *watch.WatchTable
	srv   *netsrv.NetServer
	pclnt *procclnt.ProcClnt
	done  bool
	ch    chan bool
	fsl   *fslib.FsLib
}

func MakeFsServer(root fs.Dir, addr string, fsl *fslib.FsLib,
	mkps protsrv.MkProtServer, pclnt *procclnt.ProcClnt,
	config repl.Config) *FsServer {
	fssrv := &FsServer{}
	fssrv.root = root
	fssrv.addr = addr
	fssrv.mkps = mkps
	fssrv.stats = stats.MkStats(fssrv.root)
	fssrv.st = session.MakeSessionTable(mkps, fssrv)
	fssrv.sct = sesscond.MakeSessCondTable()
	fssrv.wt = watch.MkWatchTable()
	fssrv.srv = netsrv.MakeReplicatedNetServer(fssrv, addr, false, config)
	fssrv.pclnt = pclnt
	fssrv.ch = make(chan bool)
	fssrv.fsl = fsl
	fssrv.stats.MonitorCPUUtil()
	return fssrv
}

func (fssrv *FsServer) SetFsl(fsl *fslib.FsLib) {
	fssrv.fsl = fsl
}

func (fssrv *FsServer) Root() fs.Dir {
	return fssrv.root
}

func (fssrv *FsServer) Serve() {
	// Non-intial-named services wait on the pclnt infrastructure. Initial named waits on the channel.
	if fssrv.pclnt != nil {
		if err := fssrv.pclnt.Started(proc.GetPid()); err != nil {
			debug.PrintStack()
			log.Printf("Error Started: %v", err)
		}
		if err := fssrv.pclnt.WaitEvict(proc.GetPid()); err != nil {
			debug.PrintStack()
			log.Printf("Error WaitEvict: %v", err)
		}
	} else {
		<-fssrv.ch
	}
}

func (fssrv *FsServer) Done() {
	if fssrv.pclnt != nil {
		fssrv.pclnt.Exited(proc.GetPid(), "EVICTED")
	} else {
		if !fssrv.done {
			fssrv.done = true
			fssrv.ch <- true

		}
	}
	fssrv.stats.Done()
}

func (fssrv *FsServer) MyAddr() string {
	return fssrv.srv.MyAddr()
}

func (fssrv *FsServer) GetStats() *stats.Stats {
	return fssrv.stats
}

func (fssrv *FsServer) GetWatchTable() *watch.WatchTable {
	return fssrv.wt
}

func (fssrv *FsServer) AttachTree(uname string, aname string, sessid np.Tsession) (fs.Dir, fs.CtxI) {
	return fssrv.root, MkCtx(uname, sessid, fssrv.sct)
}

func (fssrv *FsServer) Dispatch(sid np.Tsession, msg np.Tmsg) (np.Tmsg, *np.Rerror) {
	// On attach, register the new session. Otherwise, try and return an old session
	var sess *session.Session
	var ok bool
	switch msg.(type) {
	case np.Tattach:
		sess = fssrv.st.LookupInsert(sid)
	default:
		sess, ok = fssrv.st.Lookup(sid)
		// If the session doesn't exist, return an error
		if !ok {
			return nil, &np.Rerror{fmt.Sprintf("%v: no sess %v", db.GetName(), sid)}
		}
	}
	// Process the request
	switch req := msg.(type) {
	case np.Tsetfile, np.Tgetfile, np.Tcreate, np.Topen, np.Twrite, np.Tread, np.Tremove, np.Tremovefile, np.Trenameat, np.Twstat:
		// log.Printf("%p: checkLease %v %v\n", fssrv, msg.Type(), req)
		err := sess.CheckLeases(fssrv.fsl)
		if err != nil {
			return nil, &np.Rerror{err.Error()}
		}
	case np.Tlease:
		reply := &np.Ropen{}
		// log.Printf("%v: %p lease %v %v %v\n", db.GetName(), fssrv, sid, msg.Type(), req)
		if err := sess.Lease(req.Wnames, req.Qid); err != nil {
			return nil, &np.Rerror{err.Error()}
		}
		return *reply, nil
	case np.Tunlease:
		reply := &np.Ropen{}
		// log.Printf("%v: %p unlease %v %v %v\n", db.GetName(), fssrv, sid, msg.Type(), req)
		if err := sess.Unlease(req.Wnames); err != nil {
			return nil, &np.Rerror{err.Error()}
		}
		return *reply, nil
	default:
		// log.Printf("%v: %p %v %v\n", db.GetName(), fssrv, msg.Type(), req)
	}
	fssrv.stats.StatInfo().Inc(msg.Type())
	return sess.Dispatch(msg)
}

func (fssrv *FsServer) Detach(sid np.Tsession) {
	fssrv.sct.DeleteSess(sid)
	fssrv.st.Detach(sid)
}

type Ctx struct {
	uname  string
	sessid np.Tsession
	sct    *sesscond.SessCondTable
}

func MkCtx(uname string, sessid np.Tsession, sct *sesscond.SessCondTable) *Ctx {
	return &Ctx{uname, sessid, sct}
}

func (ctx *Ctx) Uname() string {
	return ctx.uname
}

func (ctx *Ctx) SessionId() np.Tsession {
	return ctx.sessid
}

func (ctx *Ctx) SessCondTable() *sesscond.SessCondTable {
	return ctx.sct
}
