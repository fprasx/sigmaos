package memfs

import (
	//"time"

	"sigmaos/file"
	"sigmaos/fs"
	"sigmaos/serr"
	sp "sigmaos/sigmap"
)

type File struct {
	fs.Inode
	*file.File
}

func MakeFile(i fs.Inode) *File {
	f := &File{}
	f.Inode = i
	f.File = file.MakeFile()
	return f
}

func (f *File) Size() (sp.Tlength, *serr.Err) {
	return f.File.Size()
}

func (f *File) Stat(ctx fs.CtxI) (*sp.Stat, *serr.Err) {
	st, err := f.Inode.Stat(ctx)
	if err != nil {
		return nil, err
	}
	l, _ := f.Size()
	st.Length = uint64(l)
	return st, nil
}

func (f *File) Snapshot(fn fs.SnapshotF) []byte {
	return makeFileSnapshot(f)
}

func RestoreFile(fn fs.RestoreF, b []byte) fs.Inode {
	return restoreFile(fn, b)
}
