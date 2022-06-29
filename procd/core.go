package procd

import (
	"runtime/debug"

	db "ulambda/debug"
	"ulambda/linuxsched"
	np "ulambda/ninep"
	"ulambda/proc"
	"ulambda/resource"
)

type Tcorestatus uint8

const (
	CORE_AVAILABLE Tcorestatus = iota
	CORE_BLOCKED               // Not for use by this procd's procs.
)

func (st Tcorestatus) String() string {
	switch st {
	case CORE_AVAILABLE:
		return "CORE_AVAILABLE"
	case CORE_BLOCKED:
		return "CORE_BLOCKED"
	default:
		db.DFatalf("Unrecognized core status")
	}
	return ""
}

func (pd *Procd) initCores(grantedCoresIv string) {
	grantedCores := np.MkInterval(0, 0)
	grantedCores.Unmarshal(grantedCoresIv)
	// First, revoke access to all cores.
	allCoresIv := np.MkInterval(0, np.Toffset(linuxsched.NCores))
	revokeMsg := resource.MakeResourceMsg(resource.Trequest, resource.Tcore, allCoresIv.String(), int(linuxsched.NCores))
	pd.removeCores(revokeMsg)

	// Then, enable access to the granted core interval.
	grantMsg := resource.MakeResourceMsg(resource.Tgrant, resource.Tcore, grantedCores.String(), int(grantedCores.Size()))
	pd.addCores(grantMsg)
}

func (pd *Procd) addCores(msg *resource.ResourceMsg) {
	cores := parseCoreInterval(msg.Name)
	pd.adjustCoresOwned(pd.coresOwned, pd.coresOwned+proc.Tcore(msg.Amount), cores, CORE_AVAILABLE)
	// Notify sleeping workers that they may be able to run procs now.
	go func() {
		for i := 0; i < msg.Amount; i++ {
			pd.stealChan <- true
		}
	}()
}

func (pd *Procd) removeCores(msg *resource.ResourceMsg) {
	cores := parseCoreInterval(msg.Name)
	pd.adjustCoresOwned(pd.coresOwned, pd.coresOwned-proc.Tcore(msg.Amount), cores, CORE_BLOCKED)
}

func (pd *Procd) adjustCoresOwned(oldNCoresOwned, newNCoresOwned proc.Tcore, coresToMark []uint, newCoreStatus Tcorestatus) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	// Mark cores according to their new status.
	pd.markCoresL(coresToMark, newCoreStatus)
	// Set the new procd core affinity.
	pd.setCoreAffinityL()
	// Rebalance procs given new cores.
	pd.rebalanceProcs(oldNCoresOwned, newNCoresOwned, coresToMark, newCoreStatus)
	pd.sanityCheckCoreCountsL()
}

// Rebalances procs across set of available cores. We allocate each proc a
// share of the owned cores proportional to their prior allocation, or the
// number of cores the proc requested, whichever is less. For simplicity, we
// currently move around all of the procs, even if they aren't having their
// cores revoked. In future, we should probably only move procs which
// absolutely need to release their cores.
func (pd *Procd) rebalanceProcs(oldNCoresOwned, newNCoresOwned proc.Tcore, coresToMark []uint, newCoreStatus Tcorestatus) {
	// Free all procs' cores.
	for _, p := range pd.runningProcs {
		pd.freeCoresL(p)
	}
	// Sanity check
	if pd.coresAvail != oldNCoresOwned {
		db.DFatalf("Mismatched num cores avail during rebalance: %v != %v", pd.coresAvail, oldNCoresOwned)
	}
	// Update the number of cores owned/available.
	pd.coresOwned = newNCoresOwned
	pd.coresAvail = newNCoresOwned
	toEvict := map[proc.Tpid]*LinuxProc{}
	// Calculate new core allocation for each proc, and allocate it cores. Some
	// of these procs may need to be evicted if there isn't enough space for
	// them.
	for pid, p := range pd.runningProcs {
		var newNCore proc.Tcore
		if p.attr.Ncore == 0 {
			// If this core didn't ask for dedicated cores, it can run on all cores.
			newNCore = newNCoresOwned
		} else {
			newNCore = p.attr.Ncore * newNCoresOwned / oldNCoresOwned
			// Don't allocate more than the number of cores this proc initially asked
			// for.
			if newNCore > p.attr.Ncore {
				// XXX This seems to me like it could lead to some fishiness when
				// growing back after a shrink. One proc may not get all of its desired
				// cores back, while some of those cores may sit idle. It is simple,
				// though, so keep it for now.
				newNCore = p.attr.Ncore
			}
		}
		// If this proc would be allocated less than one core, slate it for
		// eviction, and don't alloc any cores.
		if newNCore < 1 {
			toEvict[pid] = p
		} else {
			// Resize the proc's core allocation.
			// Allocate cores to the proc.
			pd.allocCoresL(p, newNCore)
			// Set the CPU affinity for this proc to match procd.
			p.setCpuAffinityL()
		}
	}
	// See if any of the procs to be evicted can still be squeezed in, in case
	// the "proportional allocation" strategy above left some cores unused.
	for pid, p := range toEvict {
		// If the proc fits...
		if p.attr.Ncore < pd.coresAvail {
			// Allocate cores to the proc.
			pd.allocCoresL(p, p.attr.Ncore)
			// Set the CPU affinity for this proc to match procd.
			p.setCpuAffinityL()
			delete(toEvict, pid)
		}
	}
	pd.evictProcsL(toEvict)
}

// Check if this procd has enough cores to run proc p. Caller holds lock.
func (pd *Procd) hasEnoughCores(p *proc.Proc) bool {
	db.DPrintf(db.ALWAYS, "Util1 %v", pd.GetStats().GetUtil())
	// If this is an LC proc, check that we have enough cores.
	if p.Type == proc.T_LC {
		// If we have enough cores to run this job...
		if pd.coresAvail >= p.Ncore {
			return true
		}
	} else {
		// Otherwise, determine whether or not we can run the proc based on
		// utilization.
		// TODO Parametrize
		// TODO: back off to avoid taking way too many at once.
		// If utilization is below a certain threshold, take the proc.
		db.DPrintf(db.ALWAYS, "Util2 %v", pd.GetStats().GetUtil())
		if pd.GetStats().GetUtil() < 100.0 {
			return true
		}
	}
	return false
}

// Allocate cores to a proc. Caller holds lock.
func (pd *Procd) allocCoresL(p *LinuxProc, n proc.Tcore) {
	p.coresAlloced = n
	pd.coresAvail -= n
	pd.sanityCheckCoreCountsL()
}

// Set the status of a set of cores. Caller holds lock.
func (pd *Procd) markCoresL(cores []uint, status Tcorestatus) {
	for _, i := range cores {
		// If we are double-setting a core's status, it's probably a bug.
		if pd.coreBitmap[i] == status {
			debug.PrintStack()
			db.DFatalf("Error (noded:%v): Double-marked cores %v == %v", proc.GetNodedId(), pd.coreBitmap[i], status)
		}
		pd.coreBitmap[i] = status
	}
}

func (pd *Procd) freeCores(p *LinuxProc) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	pd.freeCoresL(p)
}

// Free a set of cores which was being used by a proc.
func (pd *Procd) freeCoresL(p *LinuxProc) {
	// If no cores were exclusively allocated to this proc, return immediately.
	if p.attr.Ncore == proc.C_DEF {
		return
	}

	pd.coresAvail += p.coresAlloced
	p.coresAlloced = 0
	pd.sanityCheckCoreCountsL()
}

func parseCoreInterval(ivStr string) []uint {
	iv := np.MkInterval(0, 0)
	iv.Unmarshal(ivStr)
	cores := make([]uint, iv.Size())
	for i := uint(0); i < uint(iv.Size()); i++ {
		cores[i] = uint(iv.Start) + i
	}
	return cores
}

// Run a sanity check for our core resource accounting. Caller holds lock.
func (pd *Procd) sanityCheckCoreCountsL() {
	if pd.coresOwned > proc.Tcore(linuxsched.NCores) {
		db.DFatalf("Own more procd cores than there are cores on this machine: %v > %v", pd.coresOwned, linuxsched.NCores)
	}
	if pd.coresOwned < 0 {
		db.DFatalf("Own too few cores: %v <= 0", pd.coresOwned)
	}
	if pd.coresAvail < 0 {
		db.DFatalf("Too few cores available: %v < 0", pd.coresAvail)
	}
	if pd.coresAvail > pd.coresOwned {
		db.DFatalf("More cores available than cores owned: %v > %v", pd.coresAvail, pd.coresOwned)
	}
}
