package schedd

import (
	"path"
	"sync"
	"sync/atomic"
	"time"

	db "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/fslib"
	lcproto "sigmaos/lcschedsrv/proto"
	"sigmaos/linuxsched"
	"sigmaos/mem"
	"sigmaos/memfssrv"
	"sigmaos/perf"
	"sigmaos/proc"
	"sigmaos/procmgr"
	"sigmaos/procqclnt"
	"sigmaos/rpcclnt"
	"sigmaos/schedd/proto"
	"sigmaos/scheddclnt"
	sp "sigmaos/sigmap"
	"sigmaos/sigmasrv"
)

type Schedd struct {
	mu         sync.Mutex
	cond       *sync.Cond
	pmgr       *procmgr.ProcMgr
	scheddclnt *scheddclnt.ScheddClnt
	procqclnt  *procqclnt.ProcQClnt
	mcpufree   proc.Tmcpu
	memfree    proc.Tmem
	kernelId   string
	realms     []sp.Trealm
	mfs        *memfssrv.MemFs
	nProcsRun  uint64
}

func NewSchedd(mfs *memfssrv.MemFs, kernelId string, reserveMcpu uint) *Schedd {
	sd := &Schedd{
		pmgr:     procmgr.NewProcMgr(mfs, kernelId),
		realms:   make([]sp.Trealm, 0),
		mcpufree: proc.Tmcpu(1000*linuxsched.GetNCores() - reserveMcpu),
		memfree:  mem.GetTotalMem(),
		kernelId: kernelId,
		mfs:      mfs,
	}
	sd.cond = sync.NewCond(&sd.mu)
	sd.scheddclnt = scheddclnt.NewScheddClnt(mfs.SigmaClnt().FsLib)
	sd.procqclnt = procqclnt.NewProcQClnt(mfs.SigmaClnt().FsLib)
	return sd
}

func (sd *Schedd) ForceRun(ctx fs.CtxI, req proto.ForceRunRequest, res *proto.ForceRunResponse) error {
	atomic.AddUint64(&sd.nProcsRun, 1)
	p := proc.NewProcFromProto(req.ProcProto)
	// If this proc's memory has not been accounted for (it was not spawned via
	// the ProcQ), account for it.
	if !req.MemAccountedFor {
		sd.allocMem(p.GetMem())
	}
	db.DPrintf(db.SCHEDD, "[%v] %v ForceRun %v", p.GetRealm(), sd.kernelId, p.GetPid())
	start := time.Now()
	// Run the proc
	sd.spawnAndRunProc(p)
	db.DPrintf(db.SPAWN_LAT, "[%v] Schedd.ForceRun internal latency: %v", p.GetPid(), time.Since(start))
	db.DPrintf(db.SCHEDD, "[%v] %v ForceRun done %v", p.GetRealm(), sd.kernelId, p.GetPid())
	return nil
}

// Wait for a proc to mark itself as started.
func (sd *Schedd) WaitStart(ctx fs.CtxI, req proto.WaitRequest, res *proto.WaitResponse) error {
	db.DPrintf(db.SCHEDD, "WaitStart %v", req.PidStr)
	sd.pmgr.WaitStart(sp.Tpid(req.PidStr))
	db.DPrintf(db.SCHEDD, "WaitStart done %v", req.PidStr)
	return nil
}

// Wait for a proc to mark itself as started.
func (sd *Schedd) Started(ctx fs.CtxI, req proto.NotifyRequest, res *proto.NotifyResponse) error {
	db.DPrintf(db.SCHEDD, "Started %v", req.PidStr)
	start := time.Now()
	sd.pmgr.Started(sp.Tpid(req.PidStr))
	db.DPrintf(db.SPAWN_LAT, "[%v] Schedd.Started internal latency: %v", req.PidStr, time.Since(start))
	return nil
}

// Wait for a proc to be evicted.
func (sd *Schedd) WaitEvict(ctx fs.CtxI, req proto.WaitRequest, res *proto.WaitResponse) error {
	db.DPrintf(db.SCHEDD, "WaitEvict %v", req.PidStr)
	sd.pmgr.WaitEvict(sp.Tpid(req.PidStr))
	db.DPrintf(db.SCHEDD, "WaitEvict done %v", req.PidStr)
	return nil
}

// Wait for a proc to mark itself as exited.
func (sd *Schedd) Evict(ctx fs.CtxI, req proto.NotifyRequest, res *proto.NotifyResponse) error {
	db.DPrintf(db.SCHEDD, "Evict %v", req.PidStr)
	sd.pmgr.Evict(sp.Tpid(req.PidStr))
	return nil
}

// Wait for a proc to mark itself as exited.
func (sd *Schedd) WaitExit(ctx fs.CtxI, req proto.WaitRequest, res *proto.WaitResponse) error {
	db.DPrintf(db.SCHEDD, "WaitExit %v", req.PidStr)
	res.Status = sd.pmgr.WaitExit(sp.Tpid(req.PidStr))
	db.DPrintf(db.SCHEDD, "WaitExit done %v", req.PidStr)
	return nil
}

// Wait for a proc to mark itself as exited.
func (sd *Schedd) Exited(ctx fs.CtxI, req proto.NotifyRequest, res *proto.NotifyResponse) error {
	db.DPrintf(db.SCHEDD, "Exited %v", req.PidStr)
	sd.pmgr.Exited(sp.Tpid(req.PidStr), req.Status)
	return nil
}

// Get CPU shares assigned to this realm.
func (sd *Schedd) GetCPUShares(ctx fs.CtxI, req proto.GetCPUSharesRequest, res *proto.GetCPUSharesResponse) error {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	sm := sd.pmgr.GetCPUShares()
	smap := make(map[string]int64, len(sm))
	for r, s := range sm {
		smap[r.String()] = int64(s)
	}
	res.Shares = smap
	return nil
}

// Get schedd's CPU util.
func (sd *Schedd) GetCPUUtil(ctx fs.CtxI, req proto.GetCPUUtilRequest, res *proto.GetCPUUtilResponse) error {
	res.Util = sd.pmgr.GetCPUUtil(sp.Trealm(req.RealmStr))
	return nil
}

// For resource accounting purposes, it is assumed that only one getQueuedProcs
// thread runs per schedd.
func (sd *Schedd) getQueuedProcs() {
	// If true, bias choice of procq to this schedd's kernel.
	var bias bool = true
	for {
		memFree, ok := sd.shouldGetProc()
		if !ok {
			// If no memory is available, wait for some more.
			sd.waitForMoreMem()
			continue
		}
		db.DPrintf(db.SCHEDD, "[%v] Try get proc from procq, bias=%v", sd.kernelId, bias)
		start := time.Now()
		// Try to get a proc from the proc queue.
		procMem, qlen, ok, err := sd.procqclnt.GetProc(sd.kernelId, memFree, bias)
		db.DPrintf(db.SPAWN_LAT, "GetProc latency: %v", time.Since(start))
		if err != nil {
			db.DPrintf(db.SCHEDD_ERR, "Error GetProc: %v", err)
			// If previously biased to this schedd's kernel, and GetProc returned an
			// error, then un-bias.
			//
			// If not biased to this schedd's kernel, and GetProc returned an error,
			// then bias on the next attempt.
			if bias {
				bias = false
			} else {
				bias = true
			}
			continue
		}
		if !ok {
			db.DPrintf(db.SCHEDD, "[%v] No proc on procq, try another, bias=%v qlen=%v", sd.kernelId, bias, qlen)
			// If already biased to this schedd's kernel, and no proc was available,
			// try another.
			//
			// If not biased to this schedd's kernel, and no proc was available, then
			// bias on the next attempt.
			if bias {
				bias = false
			} else {
				bias = true
			}
			continue
		}
		// Allocate memory for the proc before this loop runs again so that
		// subsequent getProc requests carry the updated memory accounting
		// information.
		sd.allocMem(procMem)
		db.DPrintf(db.SCHEDD, "[%v] Got proc from procq, bias=%v", sd.kernelId, bias)
	}
}

func (sd *Schedd) procDone(p *proc.Proc) {
	db.DPrintf(db.SCHEDD, "Proc done %v", p)
	// Free any mem the proc was using.
	sd.freeMem(p.GetMem())
}

func (sd *Schedd) spawnAndRunProc(p *proc.Proc) {
	p.SetKernelID(sd.kernelId, false)
	sd.pmgr.Spawn(p)
	// Run the proc
	go sd.runProc(p)
}

// Run a proc via the local procd. Caller holds lock.
func (sd *Schedd) runProc(p *proc.Proc) {
	db.DPrintf(db.SCHEDD, "[%v] %v runProc %v", p.GetRealm(), sd.kernelId, p)
	sd.pmgr.RunProc(p)
	sd.procDone(p)
}

// We should always take a free proc if there is memory available.
func (sd *Schedd) shouldGetProc() (proc.Tmem, bool) {
	mem := sd.getFreeMem()
	return mem, mem > 0
}

func (sd *Schedd) register() {
	rpcc, err := rpcclnt.NewRPCClnt([]*fslib.FsLib{sd.mfs.SigmaClnt().FsLib}, path.Join(sp.LCSCHED, "~any"))
	if err != nil {
		db.DFatalf("Error lsched rpccc: %v", err)
	}
	req := &lcproto.RegisterScheddRequest{
		KernelID: sd.kernelId,
		McpuInt:  uint32(sd.mcpufree),
		MemInt:   uint32(sd.memfree),
	}
	res := &lcproto.RegisterScheddResponse{}
	if err := rpcc.RPC("LCSched.RegisterSchedd", req, res); err != nil {
		db.DFatalf("Error LCSched RegisterSchedd: %v", err)
	}
}

func (sd *Schedd) logStats() {
	for {
		time.Sleep(time.Second)
		db.DPrintf(db.ALWAYS, "Ran %v total procs", atomic.LoadUint64(&sd.nProcsRun))
	}
}

func RunSchedd(kernelId string, reserveMcpu uint) error {
	pcfg := proc.GetProcEnv()
	mfs, err := memfssrv.NewMemFs(path.Join(sp.SCHEDD, kernelId), pcfg)
	if err != nil {
		db.DFatalf("Error NewMemFs: %v", err)
	}
	sd := NewSchedd(mfs, kernelId, reserveMcpu)
	ssrv, err := sigmasrv.NewSigmaSrvMemFs(mfs, sd)
	if err != nil {
		db.DFatalf("Error PDS: %v", err)
	}
	setupMemFsSrv(ssrv.MemFs)
	setupFs(ssrv.MemFs)
	// Perf monitoring
	p, err := perf.NewPerf(pcfg, perf.SCHEDD)
	if err != nil {
		db.DFatalf("Error NewPerf: %v", err)
	}
	defer p.Done()
	go sd.getQueuedProcs()
	go sd.logStats()
	sd.register()
	ssrv.RunServer()
	return nil
}
