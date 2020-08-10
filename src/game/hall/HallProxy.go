package hall

import (
	"encoding/json"
	"game/gameProxy"
	"game/hall/UserMgr"
	"game/msgType"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"redisClient"
	"sync"
)

func NewHallProxy(module *HallModule) *HallProxy {
	proxy := new(HallProxy)
	proxy.module = module
	proxy.userMgr = UserMgr.NewHallUserMgr()
	proxy.locker = new(sync.Mutex)
	proxy.Init(proxy, &module.BaseModule, module)

	return proxy
}

type HallProxy struct {
	gameProxy.GameProxy
	module  *HallModule
	userMgr *UserMgr.HallUserMgr
	locker *sync.Mutex
}

func (this *HallProxy) AddEvents() {
}

func (this *HallProxy) RemoveEvents() {
}

func (this *HallProxy) RegisterMessages() {
	this.RegisterGO(msgType.RPC_USER_LOGIN_SUCCESS, this.onUserLoginSuccess)
	this.RegisterGO(msgType.RPC_USER_LOGOUT, this.onUserLogout)
}

func (this *HallProxy) CancelMessages() {
}

func (this *HallProxy) Destroy() {
	this.GameProxy.Destroy()
}

func (this *HallProxy) onUserLoginSuccess(session gate.Session,name string) {
	userID := uint64(session.GetUserIDInt64())
	userData := this.loadUserData(uint64(userID))
	if userData == nil {
		log.Warning("加载数据失败")
		return
	}

	if userData.RegisterTime == 0 {
		userData = UserMgr.NewHallUserData(userID,name)
	}

	//这个地方先这样，后面用HallService做单线程运行
	this.locker.Lock()
	defer this.locker.Unlock()
	this.userMgr.OnUserLogin(userData,session)

	this.userMgr.GetUser(userID).SaveData()

	log.Info("玩家%s登录大厅成功,UserID=%d,当前在线人数=%d",name,userID,this.userMgr.GetUserCount())
}

func (this *HallProxy) onUserLogout(session gate.Session) {
	userID := uint64(session.GetUserIDInt64())

	//这个地方先这样，后面用HallService做单线程运行
	this.locker.Lock()
	defer this.locker.Unlock()
	user := this.userMgr.GetUser(userID)
	if user != nil {
		user.SaveData()
		this.userMgr.OnUserLogout(userID)
	}

	log.Info("玩家下线,UserID=%d",userID)
}

func (this *HallProxy) loadUserData(userID uint64) *UserMgr.HallUserData{
	key := redisClient.GetUseRoleKey(userID)
	data, _ := redisClient.Get(key)
	if data != nil {
		userData := &UserMgr.HallUserData{}
		e := json.Unmarshal(data, userData)
		if e != nil {
			log.Error("Load user data from Redis Error:%v", e)
		}
		return userData
	}
	return nil
}

