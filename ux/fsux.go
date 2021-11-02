package fsux

import (
	"log"
	"path"
	"sync"

	"ulambda/fsclnt"
	"ulambda/fslibsrv"
	"ulambda/fssrv"
	"ulambda/named"
	np "ulambda/ninep"
	"ulambda/procinit"
	"ulambda/repl"
	usync "ulambda/sync"
	// "ulambda/seccomp"
)

type FsUx struct {
	*fssrv.FsServer
	mu    sync.Mutex
	mount string
}

func RunFsUx(mount string) {
	ip, err := fsclnt.LocalIP()
	if err != nil {
		log.Fatalf("LocalIP %v %v\n", named.UX, err)
	}
	fsux := MakeReplicatedFsUx(mount, ip+":0", procinit.GetPid(), nil)
	fsux.Serve()
}

func MakeReplicatedFsUx(mount string, addr string, pid string, config repl.Config) *FsUx {
	// seccomp.LoadFilter()  // sanity check: if enabled we want fsux to fail
	fsux := &FsUx{}
	root := makeDir([]string{mount}, np.DMDIR, nil)
	srv, fsl, err := fslibsrv.MakeReplServer(root, addr, named.UX, "ux", config)
	if err != nil {
		log.Fatalf("MakeSrvFsLib %v\n", err)
	}
	fsux.FsServer = srv
	if config == nil {
		fsuxStartCond := usync.MakeCond(fsl, path.Join(named.BOOT, pid), nil, true)
		fsuxStartCond.Destroy()
	}
	return fsux
}
