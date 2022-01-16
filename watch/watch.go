package watch

import (
	"fmt"
	"log"
	"sync"

	db "ulambda/debug"
	np "ulambda/ninep"
)

type Watch struct {
	sid    np.Tsession
	cond   *sync.Cond
	closed bool
}

func mkWatch(sid np.Tsession, l sync.Locker) *Watch {
	return &Watch{sid, sync.NewCond(l), false}
}

type Watchers struct {
	sync.Mutex
	path     string // the key in WatchTable
	nref     int    // updated under table lock
	watchers []*Watch
}

func mkWatchers(path string) *Watchers {
	w := &Watchers{}
	w.path = path
	w.watchers = make([]*Watch, 0)
	return w
}

// Caller should hold ws lock on return caller has ws lock again
func (ws *Watchers) Watch(sid np.Tsession) error {
	w := mkWatch(sid, &ws.Mutex)
	ws.watchers = append(ws.watchers, w)
	w.cond.Wait()
	db.DLPrintf("WATCH", "Watch done waiting %v p '%v'\n", ws, ws.path)
	if w.closed {
		return fmt.Errorf("Session closed %v", sid)
	}
	return nil
}

func (ws *Watchers) WakeupWatchL() {
	for _, w := range ws.watchers {
		db.DLPrintf("WATCH", "WakeupWatch %v\n", w)
		w.cond.Signal()
	}
	ws.watchers = make([]*Watch, 0)
}

func (ws *Watchers) deleteSess(sid np.Tsession) {
	ws.Lock()
	defer ws.Unlock()

	tmp := ws.watchers[:0]
	for _, w := range ws.watchers {
		if w.sid == sid {
			db.DLPrintf("WATCH", "Delete watch %v\n", w)
			w.closed = true
			w.cond.Signal()
		} else {
			tmp = append(tmp, w)
		}
	}
	ws.watchers = tmp
}

type WatchTable struct {
	sync.Mutex
	watchers map[string]*Watchers
	locked   bool
}

func MkWatchTable() *WatchTable {
	wt := &WatchTable{}
	wt.watchers = make(map[string]*Watchers)
	return wt
}

// Returns locked Watchers
// XXX Normalize paths (e.g., delete extra /) so that matches
// work for equivalent paths
func (wt *WatchTable) WatchLookupL(path []string) *Watchers {
	p := np.Join(path)

	wt.Lock()
	ws, ok := wt.watchers[p]
	if !ok {
		ws = mkWatchers(p)
		wt.watchers[p] = ws
	}
	ws.nref++ /// ensure ws won't be deleted from table
	wt.Unlock()

	ws.Lock()

	return ws
}

// Release watchers for path. Caller should have watchers locked
// through WatchLookupL().
func (wt *WatchTable) Release(ws *Watchers) {
	ws.Unlock()

	wt.Lock()
	defer wt.Unlock()

	ws.nref--

	ws1, ok := wt.watchers[ws.path]
	if !ok {
		// Another thread already deleted the entry
		return
	}

	if ws != ws1 {
		log.Fatalf("Release\n")
	}

	if ws.nref == 0 {
		delete(wt.watchers, ws.path)
	}
}

// Wakeup threads waiting for a watch on this connection.  Iterating
// through wt.wathers doesn't follow lock holder and calling
// ws.deleteSess() while holding wt lock can result in deadlock, so we
// make a copy of wt.watchers while holding the lock and then iterate
// through the copy without holding the lock.  XXX better plan?
func (wt *WatchTable) DeleteSess(sid np.Tsession) {
	// log.Printf("delete %p sess %v\n", wt, sid)

	wt.Lock()
	t := make([]*Watchers, 0, len(wt.watchers))
	for _, ws := range wt.watchers {
		t = append(t, ws)
	}
	wt.Unlock()

	for _, ws := range t {
		ws.deleteSess(sid)
	}
}
