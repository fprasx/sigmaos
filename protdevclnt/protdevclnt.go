package protdevclnt

import (
	"encoding/json"
	"fmt"

	"sigmaos/fslib"
	np "sigmaos/ninep"
)

type ProtDevClnt struct {
	*fslib.FsLib
	sid string
	fn  string
}

func MkProtDevClnt(fsl *fslib.FsLib, fn string) (*ProtDevClnt, error) {
	pdc := &ProtDevClnt{}
	pdc.FsLib = fsl
	pdc.fn = fn
	b, err := pdc.GetFile(pdc.fn + "/clone")
	if err != nil {
		return nil, fmt.Errorf("Clone err %v\n", err)
	}
	pdc.sid = "/" + string(b)
	return pdc, nil
}

func (pdc *ProtDevClnt) RPC(req []byte) ([]byte, error) {
	_, err := pdc.SetFile(pdc.fn+pdc.sid+"/data", req, np.OWRITE, 0)
	if err != nil {
		return nil, fmt.Errorf("Query err %v\n", err)
	}
	// XXX maybe the caller should use Reader
	b, err := pdc.GetFile(pdc.fn + pdc.sid + "/data")
	if err != nil {
		return nil, fmt.Errorf("Query response err %v\n", err)
	}
	return b, nil
}

func (pdc *ProtDevClnt) RPCJson(arg interface{}, res interface{}) error {
	req, err := json.Marshal(arg)
	if err != nil {
		return err
	}
	rep, err := pdc.RPC(req)
	if err != nil {
		return err
	}
	return json.Unmarshal(rep, res)
}