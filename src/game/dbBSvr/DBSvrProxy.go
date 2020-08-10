package dbBSvr

import (
	"game/gameProxy"
	"game/msgType"
	"github.com/liangdas/mqant/gate"
)

func NewDBSvrProxy(module *DBSvrModule) *DBSvrProxy {
	proxy := new(DBSvrProxy)
	proxy.module = module
	proxy.mgr = NewDBManager()
	proxy.Init(proxy, &module.BaseModule, module)
	return proxy
}

type DBSvrProxy struct {
	gameProxy.GameProxy
	module *DBSvrModule
	mgr    *DBManager
}

func (this *DBSvrProxy) AddEvents() {
}

func (this *DBSvrProxy) RemoveEvents() {
}

func (this *DBSvrProxy) RegisterMessages() {
	this.RegisterGO(msgType.RPC_LOAD_USER_INFO_FROM_DB, this.onUserLogin)
	this.RegisterGO(msgType.RPC_USER_LOGOUT, this.onUserLogout)
}

func (this *DBSvrProxy) CancelMessages() {
}

func (this *DBSvrProxy) Destroy() {
	this.GameProxy.Destroy()
}

func (this *DBSvrProxy) onUserLogin(userID int64)  (bool, error) {
	return this.mgr.onUserLogin(userID)
}

func (this *DBSvrProxy) onUserLogout(session gate.Session)  {
	this.mgr.onUserLogout(session.GetUserIDInt64())
}