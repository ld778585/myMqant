package UserMgr

import (
	"encoding/json"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"redisClient"
)

func NewHallUser(userData *HallUserData,session gate.Session) *HallUser{
	user := &HallUser{}
	user.userData = userData
	return user
}

type HallUser struct {
	userData 	*HallUserData
	session 	gate.Session
}

func (this *HallUser) SyncUserInfo()  {
}

func (this *HallUser) Logout()  {
}

func (this *HallUser) GetUserSaveData() string {
	if data, e := json.Marshal(this.userData); e == nil {
		return string(data)
	}
	return ""
}

func (this *HallUser) SaveData()  {
	key := redisClient.GetUseRoleKey(this.userData.UserID)
	b, e := redisClient.Set(key, this.GetUserSaveData())
	if !b {
		log.Error("Save User Data To Redis Error:%v", e)
	}
}

