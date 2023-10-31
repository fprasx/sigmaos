package procmgr

import (
	"sync"
	"time"

	db "sigmaos/debug"
	"sigmaos/memfssrv"
	"sigmaos/proc"
	"sigmaos/procclnt"
	"sigmaos/sigmaclnt"
	sp "sigmaos/sigmap"
	"sigmaos/uprocclnt"
)

const (
	PROC_CACHE_SZ = 500
)

type ProcMgr struct {
	sync.Mutex
	mfs            *memfssrv.MemFs
	kernelId       string
	rootsc         *sigmaclnt.SigmaClnt
	updm           *uprocclnt.UprocdMgr
	sclnts         map[sp.Trealm]*sigmaclnt.SigmaClnt
	namedIPs       map[sp.Trealm]string
	cachedProcBins map[sp.Trealm]map[string]bool
	pstate         *ProcState
}

// Manages the state and lifecycle of a proc.
func NewProcMgr(mfs *memfssrv.MemFs, kernelId string) *ProcMgr {
	mgr := &ProcMgr{
		mfs:            mfs,
		kernelId:       kernelId,
		rootsc:         mfs.SigmaClnt(),
		updm:           uprocclnt.NewUprocdMgr(mfs.SigmaClnt().FsLib, kernelId),
		sclnts:         make(map[sp.Trealm]*sigmaclnt.SigmaClnt),
		namedIPs:       make(map[sp.Trealm]string),
		cachedProcBins: make(map[sp.Trealm]map[string]bool),
		pstate:         NewProcState(),
	}
	return mgr
}

// Proc has been spawned.
func (mgr *ProcMgr) Spawn(p *proc.Proc) {
	db.DPrintf(db.SPAWN_LAT, "[%v] Schedd proc spawn time %v", p.GetPid(), time.Since(p.GetSpawnTime()))
	mgr.pstate.spawn(p)
}

func (mgr *ProcMgr) RunProc(p *proc.Proc) {
	// Set the proc's kernel ID, now that a kernel has been selected to run the
	// proc.
	p.SetKernelID(mgr.kernelId, true)
	// Set the schedd IP for the proc, so it can mount this schedd in one RPC
	// (without walking down to it).
	p.SetScheddIP(mgr.mfs.MyAddr())
	p.SetNamedIP(mgr.getNamedIP(p.GetRealm()))
	s := time.Now()
	mgr.setupProcState(p)
	db.DPrintf(db.SPAWN_LAT, "[%v] Proc state setup %v", p.GetPid(), time.Since(s))
	s = time.Now()
	mgr.downloadProc(p)
	db.DPrintf(db.SPAWN_LAT, "[%v] Binary download time %v", p.GetPid(), time.Since(s))
	mgr.runProc(p)
}

func (mgr *ProcMgr) Started(pid sp.Tpid) {
	mgr.pstate.started(pid)
}

func (mgr *ProcMgr) WaitStart(pid sp.Tpid) {
	mgr.pstate.waitStart(pid)
}

func (mgr *ProcMgr) Evict(pid sp.Tpid) {
	mgr.pstate.evict(pid)
}

func (mgr *ProcMgr) WaitEvict(pid sp.Tpid) {
	mgr.pstate.waitEvict(pid)
}

func (mgr *ProcMgr) Exited(pid sp.Tpid, status []byte) {
	mgr.pstate.exited(pid, status)
}

func (mgr *ProcMgr) WaitExit(pid sp.Tpid) []byte {
	return mgr.pstate.waitExit(pid)
}

func (mgr *ProcMgr) GetCPUShares() map[sp.Trealm]uprocclnt.Tshare {
	return mgr.updm.GetCPUShares()
}

func (mgr *ProcMgr) GetCPUUtil(realm sp.Trealm) float64 {
	return mgr.updm.GetCPUUtil(realm)
}

// Set up state to notify parent that a proc crashed.
func (mgr *ProcMgr) procCrashed(p *proc.Proc, err error) {
	// Mark the proc as exited due to a crash, and record the error exit status.
	mgr.pstate.exited(p.GetPid(), proc.NewStatusErr(err.Error(), nil).Marshal())
	db.DPrintf(db.PROCMGR_ERR, "Proc %v finished with error: %v", p, err)
	mgr.getSigmaClnt(p.GetRealm()).ExitedCrashed(p.GetPid(), p.GetProcDir(), p.GetParentDir(), proc.NewStatusErr(err.Error(), nil), p.GetHow())
}

func (mgr *ProcMgr) getNamedIP(realm sp.Trealm) string {
	mgr.Lock()
	defer mgr.Unlock()

	ip, ok := mgr.namedIPs[realm]
	if !ok {
		sc := mgr.getSigmaClntL(realm)
		ip = sc.NamedAddr()[0].Addr
		mgr.namedIPs[realm] = ip
	}
	return ip
}

func (mgr *ProcMgr) getSigmaClnt(realm sp.Trealm) *sigmaclnt.SigmaClnt {
	mgr.Lock()
	defer mgr.Unlock()

	return mgr.getSigmaClntL(realm)
}

func (mgr *ProcMgr) getSigmaClntL(realm sp.Trealm) *sigmaclnt.SigmaClnt {
	var clnt *sigmaclnt.SigmaClnt
	var ok bool
	if clnt, ok = mgr.sclnts[realm]; !ok {
		// No need to make a new client for the root realm.
		if realm == sp.ROOTREALM {
			clnt = mgr.rootsc
		} else {
			var err error
			pcfg := proc.NewDifferentRealmProcEnv(mgr.rootsc.ProcEnv(), realm)
			if clnt, err = sigmaclnt.NewSigmaClnt(pcfg); err != nil {
				db.DFatalf("Err NewSigmaClntRealm: %v", err)
			}
			// Mount KPIDS.
			procclnt.MountPids(clnt.FsLib)
		}
		mgr.sclnts[realm] = clnt
	}
	return clnt
}
