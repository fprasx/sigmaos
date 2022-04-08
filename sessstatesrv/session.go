package sessstatesrv

import (
	"fmt"
	"sync"
	"time"

	//	"github.com/sasha-s/go-deadlock"

	db "ulambda/debug"
	np "ulambda/ninep"
	"ulambda/proc"
	"ulambda/threadmgr"
)

//
// A session identifies a client across TCP connections.  For each
// session, sigmaos has a protsrv.
//
// The sess lock is to serialize requests on a session.  The calls in
// this file assume the calling wg holds the sess lock.
//

type Session struct {
	sync.Mutex
	threadmgr     *threadmgr.ThreadMgr
	conn          *np.Conn
	protsrv       np.Protsrv
	lastHeartbeat time.Time
	Sid           np.Tsession
	began         bool // true if the fssrv has already begun processing ops
	running       bool // true if the session is currently running an operation.
	closed        bool // true if the session has been closed.
	timedout      bool // for debugging
}

func makeSession(protsrv np.Protsrv, sid np.Tsession, t *threadmgr.ThreadMgr) *Session {
	sess := &Session{}
	sess.threadmgr = t
	sess.protsrv = protsrv
	sess.lastHeartbeat = time.Now()
	sess.Sid = sid
	sess.lastHeartbeat = time.Now()
	return sess
}

func (sess *Session) GetConn() *np.Conn {
	sess.Lock()
	defer sess.Unlock()
	return sess.conn
}

func (sess *Session) GetThread() *threadmgr.ThreadMgr {
	return sess.threadmgr
}

// For testing. Invoking CloseConn() will eventually cause
// sess.Close() to be called by Detach().
func (sess *Session) CloseConn() {
	sess.conn.Conn.Close()
}

// Server may call Close() several times because client may reconnect
// on a session that server has terminated and the Close() will close
// the new reply channel.  // XXX close connection?
func (sess *Session) Close() {
	sess.Lock()
	defer sess.Unlock()
	sess.closed = true
	// Close the replies channel so that writer in srvconn exits
	if sess.conn != nil {
		db.DPrintf("SESSION", "%v close replies\n", sess.Sid)
		close(sess.conn.Replies)
		sess.conn = nil
	}
}

// The conn may be nil if this is a replicated op which came through
// raft; in this case, a reply is not needed. Conn maybe also be nil
// because server closed session unilaterally.
func (sess *Session) SendConn(fc *np.Fcall) {
	sess.Lock()
	conn := sess.conn
	sess.Unlock()
	if conn != nil {
		conn.Replies <- fc
	}
}

func (sess *Session) IsClosed() bool {
	sess.Lock()
	defer sess.Unlock()
	return sess.closed
}

// Change conn if the new conn is non-nil. This may occur if, for
// example, a client starts talking to a new replica or a client
// reconnects quickly.
func (sess *Session) SetConn(conn *np.Conn) *np.Err {
	sess.Lock()
	defer sess.Unlock()
	if sess.closed {
		return np.MkErr(np.TErrClosed, fmt.Sprintf("session %v", sess.Sid))
	}
	db.DPrintf("SESSION", "%v SetConn new %v\n", sess.Sid, conn)
	sess.conn = conn
	return nil
}

func (sess *Session) heartbeat(msg np.Tmsg) {
	sess.Lock()
	defer sess.Unlock()
	db.DPrintf("SESSION", "Heartbeat sess %v msg %v %v", sess.Sid, msg.Type(), msg)
	if sess.closed {
		db.DFatalf("%v heartbeat %v on closed session %v", proc.GetName(), msg, sess.Sid)
	}
	sess.lastHeartbeat = time.Now()
}

// Indirectly timeout a session
func (sess *Session) timeout() {
	sess.Lock()
	defer sess.Unlock()
	db.DPrintf("SESSION", "timeout %v", sess.Sid)
	sess.timedout = true
}

func (sess *Session) timedOut() (bool, time.Time) {
	sess.Lock()
	defer sess.Unlock()
	// If in the middle of a running op, or this fssrv hasn't begun processing
	// ops yet, refresh the heartbeat so we don't immediately time-out when the
	// op finishes.
	if sess.running || !sess.began {
		sess.lastHeartbeat = time.Now()
		return false, sess.lastHeartbeat
	}
	return sess.timedout || time.Since(sess.lastHeartbeat).Milliseconds() > np.SESSTIMEOUTMS, sess.lastHeartbeat
}

func (sess *Session) SetRunning(r bool) {
	sess.Lock()
	defer sess.Unlock()
	sess.running = r
	// If this server is replicated, it may take a couple of seconds for the
	// replication library to start up & begin processing ops. Noting when
	// processing has started for a session helps us avoid timing-out sessions
	// until they have begun processing ops.
	sess.began = true
	// Make sure to refresh timer.
	sess.lastHeartbeat = time.Now()
}
