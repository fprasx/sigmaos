package realm

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"ulambda/config"
	db "ulambda/debug"
	"ulambda/electclnt"
	"ulambda/fslib"
	"ulambda/fslibsrv"
	"ulambda/kernel"
	"ulambda/machine"
	np "ulambda/ninep"
	"ulambda/proc"
	"ulambda/procclnt"
	"ulambda/resource"
	"ulambda/stats"
)

/* RealmMgr responsibilities:
 * - Respond to resource requests from SigmaMgr, and deallocate Nodeds.
 * - Ask for more Nodeds from SigmaMgr when load increases.
 */

const (
	NODEDS = "nodeds"
)

type RealmResourceMgr struct {
	realmId string
	// ===== Relative to the sigma named =====
	sigmaFsl *fslib.FsLib
	lock     *electclnt.ElectClnt
	*config.ConfigClnt
	memfs *fslibsrv.MemFs
	*procclnt.ProcClnt
	// ===== Relative to the realm named =====
	*fslib.FsLib
}

func MakeRealmResourceMgr(realmId string) *RealmResourceMgr {
	db.DPrintf("REALMMGR", "MakeRealmResourceMgr %v", realmId)
	m := &RealmResourceMgr{}
	m.realmId = realmId
	m.sigmaFsl = fslib.MakeFsLib(proc.GetPid().String() + "-sigmafsl")
	m.ProcClnt = procclnt.MakeProcClnt(m.sigmaFsl)
	m.ConfigClnt = config.MakeConfigClnt(m.sigmaFsl)
	m.lock = electclnt.MakeElectClnt(m.sigmaFsl, path.Join(REALM_FENCES, realmId), 0777)

	var err error
	m.memfs, err = fslibsrv.MakeMemFsFsl(path.Join(REALM_MGRS, m.realmId), m.sigmaFsl, m.ProcClnt)
	if err != nil {
		db.DFatalf("Error MakeMemFs in MakeSigmaResourceMgr: %v", err)
	}

	m.initFS()

	resource.MakeCtlFile(m.receiveResourceGrant, m.handleResourceRequest, m.memfs.Root(), np.RESOURCE_CTL)

	return m
}

func (m *RealmResourceMgr) initFS() {
	dirs := []string{NODEDS}
	for _, d := range dirs {
		if err := m.sigmaFsl.MkDir(path.Join(REALM_MGRS, m.realmId, d), 0777); err != nil {
			db.DFatalf("Error Mkdir: %v", err)
		}
	}
}
func (m *RealmResourceMgr) receiveResourceGrant(msg *resource.ResourceMsg) {
	switch msg.ResourceType {
	case resource.Tcore:
		db.DPrintf("REALMMGR", "resource.Tcore granted %v", m.realmId)
		m.growRealm()
	default:
		db.DFatalf("Unexpected resource type: %v", msg.ResourceType)
	}
}

func (m *RealmResourceMgr) handleResourceRequest(msg *resource.ResourceMsg) {
	switch msg.ResourceType {
	case resource.Trealm:
		lockRealm(m.lock, m.realmId)
		defer unlockRealm(m.lock, m.realmId)

		// On realm request, shut down & kill all nodeds.
		db.DPrintf("REALMMGR", "Realm shutdown requested %v", m.realmId)
		realmCfg, err := m.getRealmConfig()
		if err != nil {
			db.DFatalf("Error get realm config.")
		}
		for _, nodedId := range realmCfg.NodedsAssigned {
			m.deallocNoded(nodedId)
		}
	case resource.Tcore:
		lockRealm(m.lock, m.realmId)
		defer unlockRealm(m.lock, m.realmId)

		db.DPrintf("REALMMGR", "resource.Tcore requested %v", m.realmId)

		nodedId, ok := m.getLeastUtilizedNoded()

		// If no Nodeds remain...
		if !ok {
			return
		}

		db.DPrintf("REALMMGR", "least utilized node in real %v: %v", m.realmId, nodedId)
		// XXX Should we prioritize defragmentation, or try to avoid evictions?
		// Dealloc the Noded. The Noded will take care of registering itself as
		// free with the SigmaMgr.

		// Read this noded's config.
		ndCfg := &NodedConfig{}
		m.ReadConfig(path.Join(NODED_CONFIG, nodedId), ndCfg)

		if len(ndCfg.Cores) == 1 {
			db.DPrintf("REALMMGR", "Noded %v in realm %v has no cores to spare. Deallocating.", nodedId, m.realmId)
			// If this noded doesn't have any core groups to spare, deallocate it.
			m.deallocNoded(nodedId)
			db.DPrintf("REALMMGR", "Dealloced from realm %v: %v", m.realmId, nodedId)
		} else {
			cores := ndCfg.Cores[len(ndCfg.Cores)-1]
			db.DPrintf("REALMMGR", "Revoking cores %v from realm %v noded %v", cores, m.realmId, nodedId)
			// Otherwise, take some cores away.
			msg := resource.MakeResourceMsg(resource.Trequest, resource.Tcore, cores.String(), int(cores.Size()))
			if _, err := m.sigmaFsl.SetFile(path.Join(REALM_MGRS, m.realmId, NODEDS, nodedId, np.RESOURCE_CTL), msg.Marshal(), np.OWRITE, 0); err != nil {
				db.DFatalf("Error SetFile: %v", err)
			}
			db.DPrintf("REALMMGR", "Revoked cores %v from realm %v noded %v", cores, m.realmId, nodedId)
		}
	default:
		db.DFatalf("Unexpected resource type: %v", msg.ResourceType)
	}
}

// This realm has been granted cores. Now grow it. Sigmamgr must hold lock.
func (m *RealmResourceMgr) growRealm() {
	// Find a machine with free cores and claim them
	machineId, nodedId, cores, ok := m.getFreeCores(1)
	if !ok {
		db.DFatalf("Unable to get free cores to grow realm")
	}
	// If we couldn't claim cores on any machines already running a noded from
	// this realm, start a new noded.
	if nodedId == "" {
		db.DPrintf("REALMMGR", "Start a new noded for realm %v on %v with cores %v", m.realmId, machineId, cores)
		// Request the machine to start a noded.
		nodedId := m.requestNoded(machineId)
		db.DPrintf("REALMMGR", "Started noded for realm %v %v on %v", m.realmId, nodedId, machineId)
		// Allocate the noded to this realm.
		m.allocNoded(m.realmId, machineId, nodedId.String(), cores)
		db.DPrintf("REALMMGR", "Allocated %v to realm %v", nodedId, m.realmId)
	} else {
		db.DPrintf("REALMMGR", "Growing noded %v core allocation on machine %v by %v", nodedId, machineId, cores)
		// Otherwise, grant new cores to this noded.
		msg := resource.MakeResourceMsg(resource.Tgrant, resource.Tcore, cores.String(), int(cores.Size()))
		if _, err := m.sigmaFsl.SetFile(path.Join(REALM_MGRS, m.realmId, NODEDS, nodedId, np.RESOURCE_CTL), msg.Marshal(), np.OWRITE, 0); err != nil {
			db.DFatalf("Error SetFile: %v", err)
		}
	}
}

func (m *RealmResourceMgr) tryClaimCores(machineId string) (*np.Tinterval, bool) {
	cdir := path.Join(machine.MACHINES, machineId, machine.CORES)
	coreGroups, err := m.sigmaFsl.GetDir(cdir)
	if err != nil {
		db.DFatalf("Error GetDir: %v", err)
	}
	// Try to steal a core group
	for i := 0; i < len(coreGroups); i++ {
		// Read the core file.
		coreFile := path.Join(cdir, coreGroups[i].Name)
		// Claim the cores
		err = m.sigmaFsl.Remove(coreFile)
		if err == nil {
			// Cores successfully claimed.
			cores := np.MkInterval(0, 0)
			cores.Unmarshal(string(coreGroups[i].Name))
			return cores, true
		} else {
			// Unexpected error
			if !np.IsErrNotfound(err) {
				db.DFatalf("Error Remove %v", err)
			}
		}
	}
	//Cores not claimed successfully
	return nil, false
}

func (m *RealmResourceMgr) getFreeCores(nRetries int) (string, string, *np.Tinterval, bool) {
	lockRealm(m.lock, m.realmId)
	defer unlockRealm(m.lock, m.realmId)

	var machineId string
	var nodedId string
	var cores *np.Tinterval
	var ok bool
	var err error

	for i := 0; i < nRetries; i++ {
		// First, try to get cores on a machine already running a noded from this
		// realm.
		ok, err = m.sigmaFsl.ProcessDir(path.Join(REALM_MGRS, m.realmId, NODEDS), func(nd *np.Stat) (bool, error) {
			ndCfg := MakeNodedConfig()
			m.ReadConfig(path.Join(NODED_CONFIG, nd.Name), ndCfg)
			// Try to claim additional cores on the machine this noded lives on.
			if c, ok := m.tryClaimCores(ndCfg.MachineId); ok {
				cores = c
				machineId = ndCfg.MachineId
				nodedId = nd.Name
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			db.DFatalf("Error ProcessDir: %v", err)
		}
		// If successfully claimed cores, return
		if ok {
			break
		}
		// Otherwise, Try to get cores on any machine.
		ok, err = m.sigmaFsl.ProcessDir(machine.MACHINES, func(st *np.Stat) (bool, error) {
			if c, ok := m.tryClaimCores(st.Name); ok {
				cores = c
				machineId = st.Name
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			db.DFatalf("Error ProcessDir: %v", err)
		}
		// If successfully claimed cores, return
		if ok {
			break
		}
	}
	return machineId, nodedId, cores, ok
}

// Request a machine to create a new Noded)
func (m *RealmResourceMgr) requestNoded(machineId string) proc.Tpid {
	pid := proc.Tpid("noded-" + proc.GenPid().String())
	msg := resource.MakeResourceMsg(resource.Trequest, resource.Tnode, pid.String(), 1)
	if _, err := m.sigmaFsl.SetFile(path.Join(machine.MACHINES, machineId, np.RESOURCE_CTL), msg.Marshal(), np.OWRITE, 0); err != nil {
		db.DFatalf("Error SetFile in requestNoded: %v", err)
	}
	return pid
}

// Alloc a Noded to this realm.
func (m *RealmResourceMgr) allocNoded(realmId, machineId, nodedId string, cores *np.Tinterval) {
	// Update the noded's config
	ndCfg := MakeNodedConfig()
	m.ReadConfig(path.Join(NODED_CONFIG, nodedId), ndCfg)
	ndCfg.RealmId = realmId
	ndCfg.Cores = append(ndCfg.Cores, cores)
	m.WriteConfig(path.Join(NODED_CONFIG, nodedId), ndCfg)

	lockRealm(m.lock, realmId)
	defer unlockRealm(m.lock, realmId)

	// Update the realm's config
	rCfg := &RealmConfig{}
	m.ReadConfig(path.Join(REALM_CONFIG, realmId), rCfg)
	rCfg.NodedsAssigned = append(rCfg.NodedsAssigned, nodedId)
	rCfg.LastResize = time.Now()
	m.WriteConfig(path.Join(REALM_CONFIG, realmId), rCfg)
}

// Deallocate a noded from a realm.
func (m *RealmResourceMgr) deallocNoded(nodedId string) {
	// Note noded de-registration
	rCfg := &RealmConfig{}
	m.ReadConfig(path.Join(REALM_CONFIG, m.realmId), rCfg)
	db.DPrintf("REALMMGR", "Dealloc noded from realm %v, choosing from: %v", m.realmId, rCfg.NodedsAssigned)
	// Remove the noded from the list of assigned nodeds.
	for i := range rCfg.NodedsAssigned {
		if rCfg.NodedsAssigned[i] == nodedId {
			rCfg.NodedsAssigned = append(rCfg.NodedsAssigned[:i], rCfg.NodedsAssigned[i+1:]...)
			break
		}
	}
	rCfg.LastResize = time.Now()
	m.WriteConfig(path.Join(REALM_CONFIG, m.realmId), rCfg)

	ndCfg := MakeNodedConfig()
	m.ReadConfig(path.Join(NODED_CONFIG, nodedId), ndCfg)
	ndCfg.Id = nodedId
	ndCfg.RealmId = kernel.NO_REALM

	// Update the noded config file.
	m.WriteConfig(path.Join(NODED_CONFIG, nodedId), ndCfg)
}

// XXX Do I really need this?
func (m *RealmResourceMgr) getRealmConfig() (*RealmConfig, error) {
	// If the realm is being shut down, the realm config file may not be there
	// anymore. In this case, another noded is not needed.
	if _, err := m.sigmaFsl.Stat(path.Join(REALM_CONFIG, m.realmId)); err != nil && strings.Contains(err.Error(), "file not found") {
		return nil, fmt.Errorf("Realm not found")
	}
	cfg := &RealmConfig{}
	m.ReadConfig(path.Join(REALM_CONFIG, m.realmId), cfg)
	return cfg, nil
}

// Get all a realm's procd's stats
func (m *RealmResourceMgr) getRealmProcdStats(nameds []string) map[string]*stats.StatInfo {
	// XXX For now we assume all the nameds are up
	stat := make(map[string]*stats.StatInfo)
	if len(nameds) == 0 {
		return stat
	}
	m.ProcessDir(np.KPIDS, func(st *np.Stat) (bool, error) {
		// If this is a procd...
		if strings.HasPrefix(st.Name, np.PROCDREL) {
			si := kernel.GetSubsystemInfo(m.FsLib, proc.Tpid(st.Name))
			s := &stats.StatInfo{}
			err := m.GetFileJson(path.Join(np.PROCD, si.Ip, np.STATSD), s)
			if err != nil {
				db.DPrintf("REALMMGR", "Error ReadFileJson in SigmaResourceMgr.getRealmProcdStats: %v", err)
				return false, nil
			}
			stat[si.NodedId] = s
		}
		return false, nil
	})
	return stat
}

func (m *RealmResourceMgr) getRealmUtil(cfg *RealmConfig) (float64, map[string]float64) {
	// Get stats
	utilMap := make(map[string]float64)
	procdStats := m.getRealmProcdStats(cfg.NamedAddrs)
	avgUtil := 0.0
	for nodedId, stat := range procdStats {
		avgUtil += stat.Util
		utilMap[nodedId] = stat.Util
	}
	if len(procdStats) > 0 {
		avgUtil /= float64(len(procdStats))
	}
	return avgUtil, utilMap
}

func (m *RealmResourceMgr) getLeastUtilizedNoded() (string, bool) {
	// Get the realm's config
	realmCfg, err := m.getRealmConfig()
	if err != nil {
		db.DFatalf("Error getRealmConfig: %v", err)
	}

	_, procdUtils := m.getRealmUtil(realmCfg)
	db.DPrintf("REALMMGR", "searching for least utilized node in realm %v, procd utils: %v", m.realmId, procdUtils)

	if len(procdUtils) == 0 {
		return "", false
	}
	// Find least utilized procd
	min := 100.0
	minNodedId := ""
	for nodedId, util := range procdUtils {
		if min >= util {
			min = util
			minNodedId = nodedId
		}
	}
	return minNodedId, true
}

func (m *RealmResourceMgr) realmShouldGrow() bool {
	lockRealm(m.lock, m.realmId)
	defer unlockRealm(m.lock, m.realmId)

	// Get the realm's config
	realmCfg, err := m.getRealmConfig()
	if err != nil {
		db.DPrintf("REALMMGR", "Error getRealmConfig: %v", err)
		return false
	}

	// If the realm is shutting down, return
	if realmCfg.Shutdown {
		return false
	}

	// If we don't have enough noded replicas to start the realm yet, we need to
	// grow the realm.
	if len(realmCfg.NodedsAssigned) < nReplicas() {
		return true
	}

	// If we haven't finished booting, we aren't ready to start scanning/growing
	// the realm.
	if len(realmCfg.NodedsActive) < nReplicas() {
		return false
	} else {
		// If the realm just finished booting, finish initialization.
		if m.FsLib == nil {
			m.FsLib = fslib.MakeFsLibAddr(proc.GetPid().String(), realmCfg.NamedAddrs)
		}
	}

	// If we have resized too recently, return
	if time.Now().Sub(realmCfg.LastResize) < np.Conf.Realm.RESIZE_INTERVAL {
		return false
	}

	avgUtil, _ := m.getRealmUtil(realmCfg)

	if avgUtil > np.Conf.Realm.GROW_CPU_UTIL_THRESHOLD {
		return true
	}
	return false
}

func (m *RealmResourceMgr) Work() {
	db.DPrintf("REALMMGR", "Realmmgr started in realm %v", m.realmId)

	m.Started()
	go func() {
		m.WaitEvict(proc.GetPid())
		db.DPrintf("REALMMGR", "Evicted!")
		m.Exited(proc.MakeStatus(proc.StatusEvicted))
		db.DPrintf("REALMMGR", "Exited")
		os.Exit(0)
	}()

	for {
		if m.realmShouldGrow() {
			db.DPrintf("REALMMGR", "Try to grow realm %v", m.realmId)
			msg := resource.MakeResourceMsg(resource.Trequest, resource.Tcore, m.realmId, 1)
			if _, err := m.sigmaFsl.SetFile(path.Join(np.SIGMACTL), msg.Marshal(), np.OWRITE, 0); err != nil {
				db.DFatalf("Error SetFile: %v", err)
			}
		}
		// Sleep for a bit.
		time.Sleep(np.Conf.Realm.SCAN_INTERVAL)
	}
}
