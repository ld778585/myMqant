package dbBSvr

import (
	"game/define"
	"github.com/liangdas/mqant/log"
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
			userData, err := mysqlClient.GetUserData(uid, name);
			if err == nil {
				if userData != nil {
					redisClient.Set(key, string(userData))
				} else {
					redisClient.Set(key,"{}")
				}
			} else {
				log.Info("onUserLogin error=%s",err.Error())
			}
		}
	}
	return true , nil
}

func (this *DBManager) onUserLogout(uid int64) {
	this.locker.Lock()
	this.mapUserID[uid] = time.Now().Unix()
	this.locker.Unlock()
	this.saveUserData(uid)
}

func (this *DBManager) saveUserData(uid int64)  {
	time.Sleep(time.Second * 1)
	for _, name := range define.USER_DATA_NAME {
		var key = redisClient.GetUserDataKey(name, int64(uid))
		data, _ := redisClient.Get(key)
		if data != nil {
			if err := mysqlClient.SaveUserData(uid, name, data); err != nil {
				log.Error("save user data to mysql ERROR:%v", err)
			} else {
				redisClient.Del(key)
			}
		}
	}
}

