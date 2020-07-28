package UserMgr

import (
	"game/hall"
)

func NewHallUserMgr(proxy *hall.HallProxy) *HallUserMgr {
	mgr := new(HallUserMgr)
	mgr.proxy = proxy
	return mgr
}

type HallUserMgr struct {
	proxy *hall.HallProxy
}
