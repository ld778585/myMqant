package UserMgr

import (
	"github.com/liangdas/mqant/gate"
)

func NewHallUserMgr() *HallUserMgr {
	mgr := new(HallUserMgr)
	mgr.mapUser = make(map[uint64]*HallUser)
	return mgr
}

type HallUserMgr struct {
	mapUser map[uint64]*HallUser
}

func (this *HallUserMgr) GetUser(userID uint64) *HallUser {
	user, _ := this.mapUser[userID]
	return user
}

func (this *HallUserMgr) GetUserCount() uint32 {
	return uint32(len(this.mapUser))
}

func (this *HallUserMgr) delUser(userID uint64) {
	delete(this.mapUser, userID)
}

func (this *HallUserMgr) OnUserLogin(userData *HallUserData, session gate.Session) bool {
	user := this.GetUser(userData.UserID)
	if user != nil && user.session != nil {
		user.session.UnBind()
		user.session.Close()
		user.session = session
		user.SyncUserInfo()
		return true
	}

	hallUser := NewHallUser(userData, session)
	if hallUser == nil {
		return false
	}

	this.mapUser[userData.UserID] = hallUser
	hallUser.SyncUserInfo()

	return true
}

func (this *HallUserMgr) OnUserLogout(userID uint64) {
	user := this.GetUser(userID)
	if user == nil {
		return
	}

	user.Logout()
	this.delUser(userID)
}
