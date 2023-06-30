package fsetcd

import (
	"strconv"
	"strings"

	"go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"

	db "sigmaos/debug"
	"sigmaos/serr"
	"sigmaos/sessp"
	sp "sigmaos/sigmap"
)

func key2path(key string) sessp.Tpath {
	parts := strings.Split(key, ":")
	p, err := strconv.ParseUint(parts[1], 16, 64)
	if err != nil {
		db.DFatalf("ParseUint %v err %v\n", key, err)
	}
	return sessp.Tpath(p)
}

func marshalDirInfo(dir *DirInfo) ([]byte, *serr.Err) {
	d := &EtcdDir{Ents: make([]*EtcdDirEnt, dir.Ents.Len())}
	idx := 0
	dir.Ents.Iter(func(name string, i interface{}) bool {
		di := i.(DirEntInfo)
		d.Ents[idx] = &EtcdDirEnt{Name: name, Path: uint64(di.Path), Perm: uint32(di.Perm)}
		idx += 1
		return true
	})
	return marshalDir(d, dir.Perm)
}

func marshalDir(dir *EtcdDir, dperm sp.Tperm) ([]byte, *serr.Err) {
	d, err := proto.Marshal(dir)
	if err != nil {
		return nil, serr.MkErrError(err)
	}
	nfd := &EtcdFile{Perm: uint32(dperm), Data: d, ClientId: uint64(sp.NoClntId)}
	b, err := proto.Marshal(nfd)
	if err != nil {
		return nil, serr.MkErrError(err)
	}
	return b, nil
}

func UnmarshalDir(b []byte) (*EtcdDir, *serr.Err) {
	dir := &EtcdDir{}
	if err := proto.Unmarshal(b, dir); err != nil {
		return nil, serr.MkErrError(err)
	}
	return dir, nil
}

func (dir *EtcdDir) lookup(name string) (*EtcdDirEnt, bool) {
	for _, e := range dir.Ents {
		if e.Name == name {
			return e, true
		}
	}
	return nil, false
}

func MkEtcdFile(perm sp.Tperm, cid sp.TclntId, data []byte) *EtcdFile {
	return &EtcdFile{Perm: uint32(perm), Data: data, ClientId: uint64(cid)}
}

// Make empty file or directory
func MkEtcdFileDir(perm sp.Tperm, path sessp.Tpath, cid sp.TclntId) (*EtcdFile, error) {
	var fdata []byte
	perm = perm | 0777
	if perm.IsDir() {
		nd := &EtcdDir{}
		nd.Ents = append(nd.Ents, &EtcdDirEnt{Name: ".", Path: uint64(path), Perm: uint32(perm)})
		d, err := proto.Marshal(nd)
		if err != nil {
			return nil, err
		}
		fdata = d
	}
	return MkEtcdFile(perm, cid, fdata), nil
}

func (nf *EtcdFile) Tperm() sp.Tperm {
	return sp.Tperm(nf.Perm)
}

func (nf *EtcdFile) TclntId() sp.TclntId {
	return sp.TclntId(nf.ClientId)
}

func (nf *EtcdFile) TLeaseID() clientv3.LeaseID {
	return clientv3.LeaseID(nf.LeaseId)
}

func (nf *EtcdFile) SetLeaseId(lid clientv3.LeaseID) {
	nf.LeaseId = int64(lid)
}

func (e *EtcdDirEnt) Tpath() sessp.Tpath {
	return sessp.Tpath(e.Path)
}

func (e *EtcdDirEnt) Tperm() sp.Tperm {
	return sp.Tperm(e.Perm)
}