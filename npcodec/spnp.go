package npcodec

import (
	np "sigmaos/ninep"
	"sigmaos/serr"
	"sigmaos/sessp"
	sp "sigmaos/sigmap"
)

type Fcall9P struct {
	Type sessp.Tfcall
	Tag  sessp.Ttag
	Msg  sessp.Tmsg
}

func sp2NpQid(spqid sp.Tqid) np.Tqid9P {
	npqid := np.Tqid9P{}
	npqid.Type = np.Qtype9P(spqid.Type)
	npqid.Version = np.TQversion(spqid.Version)
	npqid.Path = np.Tpath(spqid.Path)
	return npqid
}

func np2SpQid(npqid np.Tqid9P) *sp.Tqid {
	spqid := &sp.Tqid{}
	spqid.Type = uint32(npqid.Type)
	spqid.Version = uint32(npqid.Version)
	spqid.Path = uint64(npqid.Path)
	return spqid
}

func Sp2NpStat(spst *sp.Stat) *np.Stat9P {
	npst := &np.Stat9P{}
	npst.Type = uint16(spst.Type)
	npst.Dev = spst.Dev
	npst.Qid = sp2NpQid(*spst.Qid)
	npst.Mode = np.Tperm(spst.Mode)
	npst.Atime = spst.Atime
	npst.Mtime = spst.Mtime
	npst.Length = np.Tlength(spst.Length)
	npst.Name = spst.Name
	npst.Uid = spst.Uid
	npst.Gid = spst.Gid
	npst.Muid = spst.Muid
	return npst
}

func Np2SpStat(npst np.Stat9P) *sp.Stat {
	spst := &sp.Stat{}
	spst.Type = uint32(npst.Type)
	spst.Dev = npst.Dev
	spst.Qid = np2SpQid(npst.Qid)
	spst.Mode = uint32(npst.Mode)
	spst.Atime = npst.Atime
	spst.Mtime = npst.Mtime
	spst.Length = uint64(npst.Length)
	spst.Name = npst.Name
	spst.Uid = npst.Uid
	spst.Gid = npst.Gid
	spst.Muid = npst.Muid
	return spst
}

func to9P(fm *sessp.FcallMsg) *Fcall9P {
	fcall9P := &Fcall9P{}
	fcall9P.Type = sessp.Tfcall(fm.Fc.Type)
	fcall9P.Tag = sessp.Ttag(fm.Fc.Tag)
	fcall9P.Msg = fm.Msg
	return fcall9P
}

func toSP(fcall9P *Fcall9P) *sessp.FcallMsg {
	fm := sessp.MakeFcallMsgNull()
	fm.Fc.Type = uint32(fcall9P.Type)
	fm.Fc.Tag = uint32(fcall9P.Tag)
	fm.Fc.Session = uint64(sessp.NoSession)
	fm.Fc.Seqno = uint64(sessp.NoSeqno)
	fm.Msg = fcall9P.Msg
	return fm
}

func np2SpMsg(fcm *sessp.FcallMsg) {
	switch fcm.Type() {
	case sessp.TTread:
		m := fcm.Msg.(*np.Tread)
		r := sp.MkReadV(sp.Tfid(m.Fid), sp.Toffset(m.Offset), sessp.Tsize(m.Count), 0)
		fcm.Msg = r
	case sessp.TTwrite:
		m := fcm.Msg.(*np.Twrite)
		r := sp.MkTwriteV(sp.Tfid(m.Fid), sp.Toffset(m.Offset), 0)
		fcm.Msg = r
		fcm.Data = m.Data
	case sessp.TTopen9P:
		m := fcm.Msg.(*np.Topen9P)
		r := sp.MkTopen(sp.Tfid(m.Fid), sp.Tmode(m.Mode))
		fcm.Msg = r
	case sessp.TTcreate9P:
		m := fcm.Msg.(*np.Tcreate9P)
		r := sp.MkTcreate(sp.Tfid(m.Fid), m.Name, sp.Tperm(m.Perm), sp.Tmode(m.Mode))
		fcm.Msg = r
	case sessp.TTwstat9P:
		m := fcm.Msg.(*np.Twstat9P)
		r := sp.MkTwstat(sp.Tfid(m.Fid), Np2SpStat(m.Stat))
		fcm.Msg = r
	}
}

func sp2NpMsg(fcm *sessp.FcallMsg) {
	switch fcm.Type() {
	case sessp.TRread:
		fcm.Fc.Type = uint32(sessp.TRread9P)
		fcm.Msg = np.Rread9P{Data: fcm.Data}
		fcm.Data = nil
	case sessp.TRerror:
		fcm.Fc.Type = uint32(sessp.TRerror9P)
		m := fcm.Msg.(*sp.Rerror)
		fcm.Msg = np.Rerror9P{Ename: serr.Terror(m.ErrCode).String()}
	}

}
