package sesssrv

import (
	"log"
	"reflect"
	"runtime/debug"

	"ulambda/ctx"
	db "ulambda/debug"
	"ulambda/dir"
	"ulambda/fencefs"
	"ulambda/fs"
	"ulambda/fslib"
	"ulambda/netsrv"
	np "ulambda/ninep"
	"ulambda/overlay"
	"ulambda/proc"
	"ulambda/procclnt"
	"ulambda/repl"
	"ulambda/sesscond"
	"ulambda/session"
	"ulambda/snapshot"
	"ulambda/stats"
	"ulambda/threadmgr"
	"ulambda/watch"
)

//
// There is one SessSrv per server. The SessSrv has one protsrv per
// session (i.e., TCP connection). Each session may multiplex several
// users.
//
// SessSrv has a table with all sess conds in use so that it can
// unblock threads that are waiting in a sess cond when a session
// closes.
//

type SessSrv struct {
	addr       string
	root       fs.Dir
	mkps       np.MkProtServer
	rps        np.RestoreProtServer
	stats      *stats.Stats
	st         *session.SessionTable
	sm         *session.SessionMgr
	sct        *sesscond.SessCondTable
	tmt        *threadmgr.ThreadMgrTable
	wt         *watch.WatchTable
	ffs        fs.Dir
	srv        *netsrv.NetServer
	replSrv    repl.Server
	rc         *repl.ReplyCache
	pclnt      *procclnt.ProcClnt
	snap       *snapshot.Snapshot
	done       bool
	replicated bool
	ch         chan bool
	fsl        *fslib.FsLib
}

func MakeSessSrv(root fs.Dir, addr string, fsl *fslib.FsLib,
	mkps np.MkProtServer, rps np.RestoreProtServer, pclnt *procclnt.ProcClnt,
	config repl.Config) *SessSrv {
	ssrv := &SessSrv{}
	ssrv.replicated = config != nil && !reflect.ValueOf(config).IsNil()
	dirover := overlay.MkDirOverlay(root)
	ssrv.root = dirover
	ssrv.addr = addr
	ssrv.mkps = mkps
	ssrv.rps = rps
	ssrv.stats = stats.MkStatsDev(ssrv.root)
	ssrv.tmt = threadmgr.MakeThreadMgrTable(ssrv.process, ssrv.replicated)
	ssrv.st = session.MakeSessionTable(mkps, ssrv, ssrv.tmt)
	ssrv.sm = session.MakeSessionMgr(ssrv.st, ssrv.SrvFcall)
	ssrv.sct = sesscond.MakeSessCondTable(ssrv.st)
	ssrv.wt = watch.MkWatchTable(ssrv.sct)
	ssrv.srv = netsrv.MakeNetServer(ssrv, addr)
	ssrv.rc = repl.MakeReplyCache()

	// Build up overlay directory
	ssrv.ffs = fencefs.MakeRoot(ctx.MkCtx("", 0, nil))

	dirover.Mount(np.STATSD, ssrv.stats)
	dirover.Mount(np.FENCEDIR, ssrv.ffs.(*dir.DirImpl))

	if !ssrv.replicated {
		ssrv.replSrv = nil
	} else {
		snapDev := snapshot.MakeDev(ssrv, nil, ssrv.root)
		dirover.Mount(np.SNAPDEV, snapDev)

		ssrv.replSrv = config.MakeServer(ssrv.tmt.AddThread())
		ssrv.replSrv.Start()
		log.Printf("Starting repl server")
	}
	ssrv.pclnt = pclnt
	ssrv.ch = make(chan bool)
	ssrv.fsl = fsl
	ssrv.stats.MonitorCPUUtil()
	return ssrv
}

func (ssrv *SessSrv) SetFsl(fsl *fslib.FsLib) {
	ssrv.fsl = fsl
}

func (ssrv *SessSrv) GetSessCondTable() *sesscond.SessCondTable {
	return ssrv.sct
}

func (ssrv *SessSrv) Root() fs.Dir {
	return ssrv.root
}

func (ssrv *SessSrv) Snapshot() []byte {
	log.Printf("Snapshot %v", proc.GetPid())
	if !ssrv.replicated {
		log.Fatalf("FATAL: Tried to snapshot an unreplicated server %v", proc.GetName())
	}
	ssrv.snap = snapshot.MakeSnapshot(ssrv)
	return ssrv.snap.Snapshot(ssrv.root.(*overlay.DirOverlay), ssrv.st, ssrv.tmt, ssrv.rc)
}

func (ssrv *SessSrv) Restore(b []byte) {
	if !ssrv.replicated {
		log.Fatal("FATAL: Tried to restore an unreplicated server %v", proc.GetName())
	}
	// Store snapshot for later use during restore.
	ssrv.snap = snapshot.MakeSnapshot(ssrv)
	ssrv.stats.Done()
	// XXX How do we install the sct and wt? How do we sunset old state when
	// installing a snapshot on a running server?
	ssrv.root, ssrv.ffs, ssrv.stats, ssrv.st, ssrv.tmt, ssrv.rc = ssrv.snap.Restore(ssrv.mkps, ssrv.rps, ssrv, ssrv.tmt.AddThread(), ssrv.process, ssrv.rc, b)
	ssrv.stats.MonitorCPUUtil()
	ssrv.sct.St = ssrv.st
	ssrv.sm.Stop()
	ssrv.sm = session.MakeSessionMgr(ssrv.st, ssrv.SrvFcall)
}

func (ssrv *SessSrv) Sess(sid np.Tsession) *session.Session {
	sess, ok := ssrv.st.Lookup(sid)
	if !ok {
		log.Fatalf("FATAL %v: no sess %v\n", proc.GetName(), sid)
		return nil
	}
	return sess
}

// The server using ssrv is ready to take requests. Keep serving
// until ssrv is told to stop using Done().
func (ssrv *SessSrv) Serve() {
	// Non-intial-named services wait on the pclnt infrastructure. Initial named waits on the channel.
	if ssrv.pclnt != nil {
		if err := ssrv.pclnt.Started(); err != nil {
			debug.PrintStack()
			log.Printf("%v: Error Started: %v", proc.GetName(), err)
		}
		if err := ssrv.pclnt.WaitEvict(proc.GetPid()); err != nil {
			debug.PrintStack()
			log.Printf("%v: Error WaitEvict: %v", proc.GetName(), err)
		}
	} else {
		<-ssrv.ch
	}
}

// The server using ssrv is done; exit.
func (ssrv *SessSrv) Done() {
	if ssrv.pclnt != nil {
		ssrv.pclnt.Exited(proc.MakeStatus(proc.StatusEvicted))
	} else {
		if !ssrv.done {
			ssrv.done = true
			ssrv.ch <- true

		}
	}
	ssrv.stats.Done()
}

func (ssrv *SessSrv) MyAddr() string {
	return ssrv.srv.MyAddr()
}

func (ssrv *SessSrv) GetStats() *stats.Stats {
	return ssrv.stats
}

func (ssrv *SessSrv) GetWatchTable() *watch.WatchTable {
	return ssrv.wt
}

func (ssrv *SessSrv) GetSnapshotter() *snapshot.Snapshot {
	return ssrv.snap
}

func (ssrv *SessSrv) AttachTree(uname string, aname string, sessid np.Tsession) (fs.Dir, fs.CtxI) {
	return ssrv.root, ctx.MkCtx(uname, sessid, ssrv.sct)
}

func (ssrv *SessSrv) SrvFcall(fc *np.Fcall, conn *np.Conn) {
	// The replies channel will be set here.
	sess := ssrv.st.Alloc(fc.Session, conn)
	// New thread about to start
	sess.IncThreads()
	if !ssrv.replicated {
		sess.GetThread().Process(fc)
	} else {
		ssrv.replSrv.Process(fc)
	}
}

func (ssrv *SessSrv) sendReply(request *np.Fcall, reply np.Tmsg, sess *session.Session) {
	fcall := np.MakeFcall(reply, 0, nil, np.NoFence)
	fcall.Session = request.Session
	fcall.Seqno = request.Seqno
	fcall.Tag = request.Tag
	db.DPrintf("SSRV", "Request %v start sendReply %v", request, fcall)
	// Store the reply in the reply cache.
	ssrv.rc.Put(request, fcall)
	// Only send a reply if the session hasn't been closed, or this is a detach
	// (the last reply to be sent).
	if !sess.IsClosed() {
		conn := sess.GetConn()
		// The conn may be nil if this is a replicated op which came
		// through raft. In this case, a reply is not needed.
		if conn != nil {
			conn.Replies <- fcall
		}
	}
	db.DPrintf("SSRV", "Request %v done sendReply %v", request, fcall)
}

func (ssrv *SessSrv) process(fc *np.Fcall) {
	// If this is a replicated op received through raft (not directly from a
	// client), the first time Alloc is called will be in this function, so the
	// reply channel will be set to nil. If it came from the client, the reply
	// channel will already be set.
	sess := ssrv.st.Alloc(fc.Session, nil)
	// Reply cache needs to live under the replication layer in order to
	// handle duplicate requests. These may occur if, for example:
	//
	// 1. A client connects to replica A and issues a request.
	// 2. Replica A pushes the request through raft.
	// 3. Before responding to the client, replica A crashes.
	// 4. The client connects to replica B, and retries the request *before*
	//    replica B hears about the request through raft.
	// 5. Replica B pushes the request through raft.
	// 6. Replica B now receives the same request twice through raft's apply
	//    channel, and will try to execute the request twice.
	//
	// In order to handle this, we can use the reply cache to deduplicate
	// requests. Since requests execute sequentially, one of the requests will
	// register itself first in the reply cache. The other request then just
	// has to wait on the reply future in order to send the reply. This can
	// happen asynchronously since it doesn't affect server state, and the
	// asynchrony is necessary in order to allow other ops on the thread to
	// make progress. We coulld optionally use sessconds, but they're kind of
	// overkill since we don't care about ordering in this case.
	if replyFuture, ok := ssrv.rc.Get(fc); ok {
		db.DPrintf("SSRV", "Request %v reply in cache", fc)
		go func() {
			ssrv.sendReply(fc, replyFuture.Await().GetMsg(), sess)
		}()
		return
	}
	db.DPrintf("SSRV", "Request %v reply not in cache", fc)
	// If this request has not been registered with the reply cache yet, register
	// it.
	ssrv.rc.Register(fc)
	ssrv.stats.StatInfo().Inc(fc.Msg.Type())
	ssrv.fenceFcall(sess, fc)
}

// Fence an fcall, if the call has a fence associated with it.  Note: don't fence blocking
// ops.
func (ssrv *SessSrv) fenceFcall(sess *session.Session, fc *np.Fcall) {
	db.DPrintf("FENCES", "fenceFcall %v fence %v\n", fc.Type, fc.Fence)
	if f, err := fencefs.CheckFence(ssrv.ffs, fc.Fence); err != nil {
		reply := *err.Rerror()
		ssrv.sendReply(fc, reply, sess)
		return
	} else {
		if f == nil {
			ssrv.serve(sess, fc)
		} else {
			defer f.Unlock()
			ssrv.serve(sess, fc)
		}
	}
}

func (ssrv *SessSrv) serve(sess *session.Session, fc *np.Fcall) {
	db.DPrintf("SSRV", "Dispatch request %v", fc)
	reply, rerror := sess.Dispatch(fc.Msg)
	db.DPrintf("SSRV", "Done dispatch request %v", fc)
	// We decrement the number of waiting threads if this request was made to
	// this server (it didn't come through raft), which will only be the case
	// when replies is not nil
	if sess.GetConn() != nil {
		defer sess.DecThreads()
	}
	if rerror != nil {
		reply = *rerror
	}
	// Send reply will drop the reply if the replies channel is nil, but it will
	// make sure to insert the reply into the reply cache.
	ssrv.sendReply(fc, reply, sess)
}

func (ssrv *SessSrv) PartitionClient(permanent bool) {
	if permanent {
		ssrv.sm.TimeoutSession()
	} else {
		ssrv.sm.CloseConn()
	}
}
