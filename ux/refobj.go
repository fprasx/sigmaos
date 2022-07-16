package fsux

import (
	"sync"

	"ulambda/fs"
	np "ulambda/ninep"
	"ulambda/refmap"
)

// Objects for which a client has an fid. Several clients may have an
// fid for the same object, therefore we ref count the object
// references.
type ObjTable struct {
	sync.Mutex
	*refmap.RefTable
}

func MkObjTable() *ObjTable {
	ot := &ObjTable{}
	ot.RefTable = refmap.MkRefTable()
	return ot
}

func (ot *ObjTable) GetRef(path np.Tpath) fs.FsObj {
	ot.Lock()
	defer ot.Unlock()

	if e, ok := ot.RefTable.Lookup(path); ok {
		return e.(fs.FsObj)
	}
	return nil
}

func (ot *ObjTable) AllocRef(path np.Tpath, o fs.FsObj) fs.FsObj {
	ot.Lock()
	defer ot.Unlock()
	e, _ := ot.RefTable.Insert(path, o)
	return e.(fs.FsObj)
}

func (ot *ObjTable) Clunk(p np.Tpath) {
	ot.Lock()
	defer ot.Unlock()

	ot.RefTable.Delete(p)
}
