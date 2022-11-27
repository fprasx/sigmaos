package protdevclntgrp

import (
	"strconv"

	"sigmaos/fslib"
	"sigmaos/group"
	"sigmaos/protdevclnt"
)

type ClntGroup struct {
	*fslib.FsLib
	clnts []*protdevclnt.ProtDevClnt
}

func MkProtDevClntGrp(fsl *fslib.FsLib, fn string, n int) (*ClntGroup, error) {
	clntgrp := &ClntGroup{}
	clntgrp.clnts = make([]*protdevclnt.ProtDevClnt, 0)
	clntgrp.FsLib = fsl
	for g := 0; g < n; g++ {
		gn := group.GRP + strconv.Itoa(g)
		pdc, err := protdevclnt.MkProtDevClnt(fsl, fn+gn)
		if err != nil {
			return nil, err
		}
		clntgrp.clnts = append(clntgrp.clnts, pdc)
	}
	return clntgrp, nil
}

func (gc *ClntGroup) Nshard() int {
	return len(gc.clnts)
}

func (gc *ClntGroup) RPC(g int, m string, arg any, res any) error {
	return gc.clnts[g].RPC(m, arg, res)
}