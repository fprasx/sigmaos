package npcodec

import (
	"fmt"

	np "ulambda/ninep"
)

// Adopted from https://github.com/docker/go-p9p/message.go

func newMsg(typ np.Tfcall) (np.Tmsg, error) {
	switch typ {
	case np.TTversion:
		return np.Tversion{}, nil
	case np.TRversion:
		return np.Rversion{}, nil
	case np.TTauth:
		return np.Tauth{}, nil
	case np.TRauth:
		return np.Rauth{}, nil
	case np.TTattach:
		return np.Tattach{}, nil
	case np.TRattach:
		return np.Rattach{}, nil
	case np.TRerror:
		return np.Rerror{}, nil
	case np.TTflush:
		return np.Tflush{}, nil
	case np.TRflush:
		return np.Rflush{}, nil
	case np.TTwalk:
		return np.Twalk{}, nil
	case np.TRwalk:
		return np.Rwalk{}, nil
	case np.TTopen:
		return np.Topen{}, nil
	case np.TRopen:
		return np.Ropen{}, nil
	case np.TTcreate:
		return np.Tcreate{}, nil
	case np.TRcreate:
		return np.Rcreate{}, nil
	case np.TTread:
		return np.Tread{}, nil
	case np.TRread:
		return np.Rread{}, nil
	case np.TTwrite:
		return np.Twrite{}, nil
	case np.TRwrite:
		return np.Rwrite{}, nil
	case np.TTclunk:
		return np.Tclunk{}, nil
	case np.TRclunk:
		return np.Rclunk{}, nil // no response body
	case np.TTremove:
		return np.Tremove{}, nil
	case np.TRremove:
		return np.Rremove{}, nil
	case np.TTstat:
		return np.Tstat{}, nil
	case np.TRstat:
		return np.Rstat{}, nil
	case np.TTwstat:
		return np.Twstat{}, nil
	case np.TRwstat:
		return np.Rwstat{}, nil
	case np.TTwritev:
		return np.Twritev{}, nil
	case np.TTreadv:
		return np.Treadv{}, nil
	case np.TTopenv:
		return np.Topenv{}, nil
	}
	return nil, fmt.Errorf("unknown message type")
}
