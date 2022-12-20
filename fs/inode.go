package fs

import (
	sp "sigmaos/sigmap"
    "sigmaos/sessp"
    "sigmaos/serr"
)

type SnapshotF func(Inode) sessp.Tpath
type RestoreF func(sessp.Tpath) Inode

// Inode interface for directories

type Inode interface {
	FsObj
	SetMtime(int64)
	Mtime() int64
	Size() (sp.Tlength, *serr.Err)
	SetParent(Dir)
	Unlink()
	Snapshot(SnapshotF) []byte
}
