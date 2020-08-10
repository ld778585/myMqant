package UserMgr

import "time"

type HallUserData struct {
	UserID         uint64
	Name           string
	Level          int32
	Money          int32
	GameIcon       uint32
	RegisterTime   int64
	LastLoginTime  int64
	LastLogoutTime int64
}

func NewHallUserData(userID uint64,name string) *HallUserData{
	userData := &HallUserData{}
	userData.UserID = userID
	userData.Name = name
	userData.RegisterTime = time.Now().Unix()
	return userData
}
