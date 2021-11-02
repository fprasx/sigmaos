package kv

import (
	"log"
	"os"
	"sync"
	"time"

	db "ulambda/debug"
	"ulambda/fslib"
	"ulambda/named"
	"ulambda/proc"
	"ulambda/procdep"
	"ulambda/procinit"
	"ulambda/stats"
	usync "ulambda/sync"
)

const (
	KV        = "bin/user/kv"
	KVMONLOCK = "monlock"
)

type Monitor struct {
	mu sync.Mutex
	*fslib.FsLib
	proc.ProcClnt
	kv        string
	kvmonlock *usync.Lock
}

func MakeMonitor(args []string) (*Monitor, error) {
	mo := &Monitor{}
	mo.FsLib = fslib.MakeFsLib(procinit.GetPid())
	mo.ProcClnt = procinit.MakeProcClnt(mo.FsLib, procinit.GetProcLayersMap())
	mo.kvmonlock = usync.MakeLock(mo.FsLib, KVDIR, KVMONLOCK, true)
	db.Name(procinit.GetPid())

	mo.kvmonlock.Lock()

	mo.Started(procinit.GetPid())
	return mo, nil
}

func (mo *Monitor) unlock() {
	mo.kvmonlock.Unlock()
}

func spawnBalancerPid(sched proc.ProcClnt, opcode, pid1, pid2 string) {
	t := procdep.MakeProcDep(pid2, "bin/user/balancer", []string{opcode, pid1})
	t.Env = []string{procinit.GetProcLayersString()}
	t.Dependencies = &procdep.Deps{map[string]bool{pid1: false}, nil}
	t.Type = proc.T_LC
	sched.Spawn(t)
}

func spawnBalancer(sched proc.ProcClnt, opcode, pid1 string) string {
	t := procdep.MakeProcDep(proc.GenPid(), "bin/user/balancer", []string{opcode, pid1})
	t.Env = []string{procinit.GetProcLayersString()}
	t.Dependencies = &procdep.Deps{map[string]bool{pid1: false}, nil}
	t.Type = proc.T_LC
	sched.Spawn(t)
	return t.Pid
}

func spawnKVPid(sched proc.ProcClnt, pid1 string, pid2 string) {
	t := procdep.MakeProcDep(pid1, KV, []string{""})
	t.Env = []string{procinit.GetProcLayersString()}
	t.Type = proc.T_LC
	sched.Spawn(t)
}

func SpawnKV(sched proc.ProcClnt) string {
	t := procdep.MakeProcDep(proc.GenPid(), KV, []string{""})
	t.Pid = proc.GenPid()
	t.Env = []string{procinit.GetProcLayersString()}
	t.Type = proc.T_LC
	sched.Spawn(t)
	return t.Pid
}

func runBalancerPid(sched proc.ProcClnt, opcode, pid1, pid2 string) {
	spawnBalancerPid(sched, opcode, pid1, pid2)
	status, err := sched.WaitExit(pid2)
	if err != nil || status != "OK" {
		log.Printf("runBalancer: err %v, status %v\n", err, status)
	}
}

func RunBalancer(sched proc.ProcClnt, opcode, pid1 string) {
	pid2 := spawnBalancer(sched, opcode, pid1)
	status, err := sched.WaitExit(pid2)
	if err != nil || status != "OK" {
		log.Printf("runBalancer: err %v status %v\n", err, status)
	}
}

func (mo *Monitor) grow() {
	pid1 := proc.GenPid()
	pid2 := proc.GenPid()
	spawnKVPid(mo.ProcClnt, pid1, pid2)
	runBalancerPid(mo.ProcClnt, "add", pid1, pid2)
}

func (mo *Monitor) shrink(kv string) {
	RunBalancer(mo.ProcClnt, "del", kv)
	err := mo.Remove(named.MEMFS + "/" + kv + "/")
	if err != nil {
		log.Printf("shrink: remove failed %v\n", err)
	}
}

// XXX Use load too?
func (mo *Monitor) Work() {
	defer mo.unlock() // release lock acquired in MakeMonitor()

	var conf *Config
	for true {
		c, err := readConfig(mo.FsLib, KVCONFIG)
		if err != nil {
			log.Printf("readConfig: err %v\n", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		conf = c
		break
	}

	now := time.Now().UnixNano()
	if (now-conf.Ctime)/1_000_000_000 < 1 {
		log.Printf("Monitor: skip\n")
		return
	}

	kvs := makeKvs(conf.Shards)
	log.Printf("Monitor config %v\n", kvs)

	util := float64(0)
	low := float64(100.0)
	lowkv := ""
	var lowload stats.Tload
	n := 0
	for kv, _ := range kvs.set {
		kvd := named.MEMFS + "/" + kv + "/statsd"
		sti := stats.StatInfo{}
		err := mo.ReadFileJson(kvd, &sti)
		if err != nil {
			log.Printf("ReadFileJson failed %v\n", err)
			os.Exit(1)
		}
		n += 1
		util += sti.Util
		if sti.Util < low {
			low = sti.Util
			lowkv = kv
			lowload = sti.Load
		}
		log.Printf("path %v\n", sti.SortPath()[0:3])
	}
	util = util / float64(n)
	log.Printf("monitor: avg util %.1f low %.1f kv %v %v\n", util, low, lowkv, lowload)
	if util >= stats.MAXLOAD {
		mo.grow()
	}
	if util < stats.MINLOAD && len(kvs.set) > 1 {
		mo.shrink(lowkv)
	}
}

func (mo *Monitor) Exit() {
	mo.Exited(procinit.GetPid(), "OK")
}
