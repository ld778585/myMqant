package dbBSvr

import (
	"game/define"
	"mysqlClient"
	"redisClient"
	"sync"
	"time"
)

type DBManager struct {
	mapUserID   map[int64]int64
	locker      *sync.RWMutex
	deleteDelay int64 //删除数据延迟
}

func NewDBManager() *DBManager {
	result := &DBManager{}
	result.init()
	return result
}

func (this *DBManager) init() {
	this.mapUserID = make(map[int64]int64)
	this.locker = new(sync.RWMutex)
	this.deleteDelay = 60
	mysqlClient.Initialize()
}

func (this *DBManager) onUserLogin(uid int64)  (bool, error){
	this.locker.Lock()
	delete(this.mapUserID,uid)
	this.locker.Unlock()

	for _, name := range define.USER_DATA_NAME {
		key := redisClient.GetUserDataKey(name,uid)
		data, _ := redisClient.Get(key)
		if data == nil {
			if userData, _ := mysqlClient.GetUserData(uid, name); userData != nil {
				redisClient.Set(key, string(userData))
			}
		}
	}
	return true , nil
}

func (this *DBManager) onUserLogout(uid int64) {
	this.locker.Lock()
	this.mapUserID[uid] = time.Now().Unix()
	this.locker.Unlock()
}

