package memfs

import (
	"ulambda/fs"
	"ulambda/inode"
	np "ulambda/ninep"
	"ulambda/pipe"
)

func MakeInode(ctx fs.CtxI, p np.Tperm, m np.Tmode, parent fs.Dir, mk fs.MakeDirF) (fs.Inode, *np.Err) {
	i := inode.MakeInode(ctx, p, parent)
	if p.IsDir() {
		return mk(i, MakeInode), nil
	} else if p.IsSymlink() {
		return MakeSym(i), nil
	} else if p.IsPipe() {
		return pipe.MakePipe(ctx, i), nil
	} else if p.IsFile() || p.IsEphemeral() {
		return MakeFile(i), nil
	} else {
		return nil, np.MkErr(np.TErrInval, p)
	}
}
