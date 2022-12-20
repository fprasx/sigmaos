package protdevsrv

import (
	"encoding/json"
	"fmt"
	"sync"

	db "sigmaos/debug"
	"sigmaos/sessp"
    "sigmaos/serr"
	"sigmaos/fs"
	"sigmaos/inode"
	"sigmaos/memfssrv"
	sp "sigmaos/sigmap"
)

type MethodStat struct {
	N   uint64 // number of invocations of method
	Tot int64  // tot us for this method
	Max int64
	Avg float64
}

func (ms *MethodStat) String() string {
	return fmt.Sprintf("N %d Tot %dus Max %dus Avg %.1fms", ms.N, ms.Tot, ms.Max, ms.Avg)
}

type Stats struct {
	MStats  map[string]*MethodStat
	AvgQLen float64
}

func mkStats() *Stats {
	st := &Stats{}
	st.MStats = make(map[string]*MethodStat)
	return st
}

func (st *Stats) String() string {
	s := "stats:\n methods:\n"
	for k, st := range st.MStats {
		s += fmt.Sprintf("  %s: %s\n", k, st.String())
	}
	s += fmt.Sprintf(" AvgQLen: %.3f", st.AvgQLen)
	return s
}

type StatInfo struct {
	sync.Mutex
	st  *Stats
	len uint64
}

func MakeStatInfo() *StatInfo {
	si := &StatInfo{}
	si.st = mkStats()
	return si
}

func (si *StatInfo) Stats() *Stats {
	n := uint64(0)
	for _, st := range si.st.MStats {
		n += st.N
		if st.N > 0 {
			st.Avg = float64(st.Tot) / float64(st.N) / 1000.0
		}
	}
	if n > 0 {
		si.st.AvgQLen = float64(si.len) / float64(n)
	}
	return si.st
}

func (sts *StatInfo) Stat(m string, t int64, ql int) {
	sts.Lock()
	defer sts.Unlock()
	sts.len += uint64(ql)
	st, ok := sts.st.MStats[m]
	if !ok {
		st = &MethodStat{}
		sts.st.MStats[m] = st
	}
	st.N += 1
	st.Tot += t
	if st.Max == 0 || t > st.Max {
		st.Max = t
	}
}

type statsDev struct {
	*inode.Inode
	si *StatInfo
}

func makeStatsDev(mfs *memfssrv.MemFs) (*StatInfo, *serr.Err) {
	std := &statsDev{}
	std.Inode = mfs.MakeDevInode()
	if err := mfs.MkDev(STATS, std); err != nil {
		return nil, err
	}
	std.si = MakeStatInfo()
	return std.si, nil
}

func (std *statsDev) Read(ctx fs.CtxI, off sp.Toffset, cnt sessp.Tsize, v sp.TQversion) ([]byte, *serr.Err) {
	if off > 0 {
		return nil, nil
	}

	std.si.Lock()
	defer std.si.Unlock()

	db.DPrintf(db.PROTDEVSRV, "Read stats: %v\n", std.si)
	std.si.Stats()
	b, err := json.Marshal(std.si.st)
	if err != nil {
		return nil, serr.MkErrError(err)
	}
	return b, nil
}

func (std *statsDev) Write(ctx fs.CtxI, off sp.Toffset, b []byte, v sp.TQversion) (sessp.Tsize, *serr.Err) {
	return 0, serr.MkErr(serr.TErrNotSupported, nil)
}

func (std *statsDev) Close(ctx fs.CtxI, m sp.Tmode) *serr.Err {
	db.DPrintf(db.PROTDEVSRV, "Close stats\n")
	return nil
}
