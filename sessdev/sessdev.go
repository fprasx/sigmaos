package sessdev

import (
	"sigmaos/clonedev"
	"sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/memfssrv"
	np "sigmaos/ninep"
)

const (
	DATA = "data-"
)

func DataName(fn string) string {
	return DATA + fn
}

type MkSessionF func(*memfssrv.MemFs, np.Tsession) (fs.Inode, *np.Err)

type SessDev struct {
	mfs *memfssrv.MemFs
	fn  string
	mks MkSessionF
}

func MkSessDev(mfs *memfssrv.MemFs, fn string, mks MkSessionF) error {
	fd := &SessDev{mfs, fn, mks}
	if err := clonedev.MkCloneDev(mfs, fn, fd.mkSession, fd.detachSession); err != nil {
		return err
	}
	return nil
}

// XXX clean up in case of error
func (fd *SessDev) mkSession(mfs *memfssrv.MemFs, sid np.Tsession) *np.Err {
	sess, err := fd.mks(mfs, sid)
	if err != nil {
		return err
	}
	sidn := clonedev.SidName(sid.String(), fd.fn)
	fn := sidn + "/" + DataName(fd.fn)
	debug.DPrintf("SESSDEV", "mkSession %v\n", fn)
	if err := mfs.MkDev(fn, sess); err != nil {
		debug.DPrintf("SESSDEV", "mkSession %v err %v\n", fn, err)
		return err
	}
	return nil
}

func (fd *SessDev) detachSession(sid np.Tsession) {
	sidn := clonedev.SidName(sid.String(), fd.fn)
	fn := sidn + "/" + DataName(fd.fn)
	if err := fd.mfs.Remove(fn); err != nil {
		debug.DPrintf("SESSDEV", "detachSession %v err %v\n", fn, err)
	}
}