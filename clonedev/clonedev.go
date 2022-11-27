package clonedev

import (
	db "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/inode"
	"sigmaos/memfssrv"
	np "sigmaos/ninep"
	"sigmaos/proc"
)

const (
	CLONE = "clone-"
	CTL   = "ctl"
)

type MkSessionF func(*memfssrv.MemFs, np.Tsession) *np.Err

func SidName(sid string, fn string) string {
	return sid + "-" + fn
}

func CloneName(fn string) string {
	return CLONE + fn
}

type Clone struct {
	*inode.Inode
	mfs       *memfssrv.MemFs
	mksession MkSessionF
	detach    np.DetachF
	fn        string
}

func makeClone(mfs *memfssrv.MemFs, fn string, mks MkSessionF, d np.DetachF) *np.Err {
	cl := &Clone{}
	cl.Inode = mfs.MakeDevInode()
	err := mfs.MkDev(CloneName(fn), cl) // put clone file into root dir
	if err != nil {
		return err
	}
	cl.mfs = mfs
	cl.mksession = mks
	cl.detach = d
	cl.fn = fn
	return nil
}

// XXX clean up in case of error
func (c *Clone) Open(ctx fs.CtxI, m np.Tmode) (fs.FsObj, *np.Err) {
	sid := ctx.SessionId()
	n := SidName(sid.String(), c.fn)
	db.DPrintf("CLONEDEV", "%v: Clone dir %v %v %v\n", proc.GetProgram(), c.fn, sid, n)
	if _, err := c.mfs.Create(n, np.DMDIR, np.ORDWR); err != nil {
		db.DPrintf("CLONEDEV", "%v: MkDir %v err %v\n", proc.GetProgram(), n, err)
		return nil, err
	}
	s := &session{}
	s.id = sid
	s.Inode = c.mfs.MakeDevInode()
	ctl := n + "/" + CTL
	err := c.mfs.MkDev(ctl, s)
	if err != nil {
		db.DPrintf("CLONEDEV", "%v: MkDev %v err %v\n", proc.GetProgram(), ctl, err)
		return nil, err
	}
	if err := c.mfs.RegisterDetach(c.Detach, sid); err != nil {
		db.DPrintf("CLONEDEV", "%v: RegisterDetach err %v\n", proc.GetProgram(), err)
	}
	if err := c.mksession(c.mfs, sid); err != nil {
		return nil, err
	}
	return s, nil
}

func (c *Clone) Close(ctx fs.CtxI, m np.Tmode) *np.Err {
	db.DPrintf("CLONEDEV", "%v: Close clone\n", proc.GetProgram())
	return nil
}

func (c *Clone) Detach(session np.Tsession) {
	db.DPrintf("CLONEDEV", "Detach %v\n", session)
	dir := SidName(session.String(), c.fn)
	n := dir + "/" + CTL
	if err := c.mfs.Remove(n); err != nil {
		db.DPrintf("CLONEDEV", "Remove %v err %v\n", n, err)
	}
	if c.detach != nil {
		c.detach(session)
	}
	if err := c.mfs.Remove(dir); err != nil {
		db.DPrintf("CLONEDEV", "Detach err %v\n", err)
	}
}

func MkCloneDev(mfs *memfssrv.MemFs, fn string, f MkSessionF, d np.DetachF) error {
	if err := makeClone(mfs, fn, f, d); err != nil {
		return err
	}
	return nil
}