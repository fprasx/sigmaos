package etcdclnt

import (
	"context"
	"strconv"
	"strings"

	"go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"

	db "sigmaos/debug"
	"sigmaos/path"
	"sigmaos/serr"
	"sigmaos/sessp"
	sp "sigmaos/sigmap"
)

const (
	BOOT sessp.Tpath = 0
)

func (ec *EtcdClnt) path2key(path sessp.Tpath) string {
	return string(ec.realm) + ":" + strconv.FormatUint(uint64(path), 16)
}

func key2path(key string) sessp.Tpath {
	parts := strings.Split(key, ":")
	p, err := strconv.ParseUint(parts[1], 16, 64)
	if err != nil {
		db.DFatalf("ParseUint %v err %v\n", key, err)
	}
	return sessp.Tpath(p)
}

func (ec *EtcdClnt) GetFile(p sessp.Tpath) (*NamedFile, sp.TQversion, *serr.Err) {
	resp, err := ec.Get(context.TODO(), ec.path2key(p))
	if err != nil {
		return nil, 0, serr.MkErrError(err)
	}
	db.DPrintf(db.ETCDCLNT, "GetFile %v %v\n", ec.path2key(p), resp)
	if len(resp.Kvs) != 1 {
		return nil, 0, serr.MkErr(serr.TErrNotfound, p)
	}
	nf := &NamedFile{}
	if err := proto.Unmarshal(resp.Kvs[0].Value, nf); err != nil {
		return nil, 0, serr.MkErrError(err)
	}
	db.DPrintf(db.ETCDCLNT, "GetFile %v %v\n", ec.path2key(p), nf)
	return nf, sp.TQversion(resp.Kvs[0].Version), nil
}

func (ec *EtcdClnt) PutFile(p sessp.Tpath, nf *NamedFile) *serr.Err {
	if b, err := proto.Marshal(nf); err != nil {
		return serr.MkErrError(err)
	} else {
		resp, err := ec.Put(context.TODO(), ec.path2key(p), string(b))
		if err != nil {
			return serr.MkErrError(err)
		}
		db.DPrintf(db.ETCDCLNT, "PutFile %v %v %v\n", p, nf, resp)
		return nil
	}
}

func (ec *EtcdClnt) ReadDir(p sessp.Tpath) (*NamedDir, sp.TQversion, *serr.Err) {
	db.DPrintf(db.ETCDCLNT, "readDir %v\n", p)
	nf, v, err := ec.GetFile(p)
	if err != nil {
		return nil, 0, err
	}
	dir, err := UnmarshalDir(nf.Data)
	if err != nil {
		return nil, 0, err
	}
	return dir, v, nil
}

// XXX retry
func (ec *EtcdClnt) Create(pn path.Path, dp sessp.Tpath, dir *NamedDir, dperm sp.Tperm, v sp.TQversion, p sessp.Tpath, perm sp.Tperm, nf *NamedFile) *serr.Err {
	b, err := proto.Marshal(nf)
	if err != nil {
		return serr.MkErrError(err)
	}
	d1, r := MarshalDir(dir, dperm)
	if r != nil {
		return r
	}
	// Update directory if new file/dir doesn't exist and directory
	// hasn't changed.
	cmp := []clientv3.Cmp{
		clientv3.Compare(clientv3.Version(ec.path2key(p)), "=", 0),
		clientv3.Compare(clientv3.Version(ec.path2key(dp)), "=", int64(v))}
	ops := []clientv3.Op{
		clientv3.OpPut(ec.path2key(p), string(b)),
		clientv3.OpPut(ec.path2key(dp), string(d1))}
	resp, err := ec.Txn(context.TODO()).If(cmp...).Then(ops...).Commit()
	if err != nil {
		return serr.MkErrError(err)
	}
	db.DPrintf(db.NAMEDV1, "addObj %v %v\n", p, resp)
	if !resp.Succeeded {
		return serr.MkErr(serr.TErrExists, p)
	}
	return nil
}

func (ec *EtcdClnt) Remove(d sessp.Tpath, dir *NamedDir, dperm sp.Tperm, v sp.TQversion, del sessp.Tpath) *serr.Err {
	d1, r := MarshalDir(dir, dperm)
	if r != nil {
		return r
	}
	cmp := []clientv3.Cmp{
		clientv3.Compare(clientv3.Version(ec.path2key(del)), ">", 0),
		clientv3.Compare(clientv3.Version(ec.path2key(d)), "=", int64(v))}
	ops := []clientv3.Op{
		clientv3.OpDelete(ec.path2key(del)),
		clientv3.OpPut(ec.path2key(d), string(d1))}
	resp, err := ec.Txn(context.TODO()).
		If(cmp...).Then(ops...).Commit()
	if err != nil {
		return serr.MkErrError(err)
	}
	db.DPrintf(db.NAMEDV1, "rmObj %v %v\n", del, resp)
	if !resp.Succeeded {
		return serr.MkErr(serr.TErrNotfound, del)
	}
	return nil
}

// XXX retry
func (ec *EtcdClnt) Rename(d sessp.Tpath, dir *NamedDir, dperm sp.Tperm, v sp.TQversion, del sessp.Tpath) *serr.Err {
	d1, r := MarshalDir(dir, dperm)
	if r != nil {
		return r
	}
	var cmp []clientv3.Cmp
	var ops []clientv3.Op
	if del != 0 {
		cmp = []clientv3.Cmp{
			clientv3.Compare(clientv3.Version(ec.path2key(del)), ">", 0),
			clientv3.Compare(clientv3.Version(ec.path2key(d)), "=", int64(v))}
		ops = []clientv3.Op{
			clientv3.OpDelete(ec.path2key(del)),
			clientv3.OpPut(ec.path2key(d), string(d1))}
	} else {
		cmp = []clientv3.Cmp{
			clientv3.Compare(clientv3.Version(ec.path2key(d)), "=", int64(v))}
		ops = []clientv3.Op{
			clientv3.OpPut(ec.path2key(d), string(d1))}
	}
	resp, err := ec.Txn(context.TODO()).If(cmp...).Then(ops...).Commit()
	if err != nil {
		return serr.MkErrError(err)
	}
	db.DPrintf(db.NAMEDV1, "mvObj %v %v\n", d, resp)
	if !resp.Succeeded {
		return serr.MkErr(serr.TErrNotfound, d)
	}
	return nil
}

// XXX retry
func (ec *EtcdClnt) RenameAt(df sessp.Tpath, dirf *NamedDir, dirfperm sp.Tperm, vf sp.TQversion, dt sessp.Tpath, dirt *NamedDir, dirtperm sp.Tperm, vt sp.TQversion, del sessp.Tpath) *serr.Err {
	bf, r := MarshalDir(dirf, dirfperm)
	if r != nil {
		return r
	}
	bt, r := MarshalDir(dirt, dirtperm)
	if r != nil {
		return r
	}
	var cmp []clientv3.Cmp
	var ops []clientv3.Op
	if del != 0 {
		cmp = []clientv3.Cmp{
			clientv3.Compare(clientv3.Version(ec.path2key(del)), ">", 0),
			clientv3.Compare(clientv3.Version(ec.path2key(df)), "=", int64(vf)),
			clientv3.Compare(clientv3.Version(ec.path2key(dt)), "=", int64(vt)),
		}
		ops = []clientv3.Op{
			clientv3.OpDelete(ec.path2key(del)),
			clientv3.OpPut(ec.path2key(df), string(bf)),
			clientv3.OpPut(ec.path2key(dt), string(bt)),
		}
	} else {
		cmp = []clientv3.Cmp{
			clientv3.Compare(clientv3.Version(ec.path2key(df)), "=", int64(vf)),
			clientv3.Compare(clientv3.Version(ec.path2key(dt)), "=", int64(vt)),
		}
		ops = []clientv3.Op{
			clientv3.OpPut(ec.path2key(df), string(bf)),
			clientv3.OpPut(ec.path2key(dt), string(bt)),
		}
	}
	resp, err := ec.Txn(context.TODO()).If(cmp...).Then(ops...).Commit()
	if err != nil {
		return serr.MkErrError(err)
	}
	db.DPrintf(db.NAMEDV1, "mvObjAt %v %v\n", del, resp)
	if !resp.Succeeded {
		return serr.MkErr(serr.TErrNotfound, del)
	}
	return nil
}

func MarshalDir(dir *NamedDir, dperm sp.Tperm) ([]byte, *serr.Err) {
	d, err := proto.Marshal(dir)
	if err != nil {
		return nil, serr.MkErrError(err)
	}
	nfd := &NamedFile{Perm: uint32(dperm), Data: d}
	b, err := proto.Marshal(nfd)
	if err != nil {
		return nil, serr.MkErrError(err)
	}
	return b, nil
}

func UnmarshalDir(b []byte) (*NamedDir, *serr.Err) {
	dir := &NamedDir{}
	if err := proto.Unmarshal(b, dir); err != nil {
		return nil, serr.MkErrError(err)
	}
	return dir, nil
}
