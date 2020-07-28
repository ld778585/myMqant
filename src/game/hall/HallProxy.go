package hall

import (
	"game/gameProxy"
	"game/hall/UserMgr"
	"game/msgType"
	"github.com/liangdas/mqant/gate"
)

func NewHallProxy(module *HallModule) *HallProxy {
	proxy := new(HallProxy)
	proxy.module = module
	proxy.userMgr = UserMgr.NewHallUserMgr(proxy)
	proxy.Init(proxy, &module.BaseModule, module)
	return proxy
}

type HallProxy struct {
	gameProxy.GameProxy
	module  *HallModule
	userMgr *UserMgr.HallUserMgr
}

func (this *HallProxy) AddEvents() {
}

func (this *HallProxy) RemoveEvents() {
}

func (this *HallProxy) RegisterMessages() {
	this.RegisterGO(msgType.RPC_USER_LOGIN_SUCCESS, this.onUserLoginSuccess)
}

func (this *HallProxy) CancelMessages() {
}

func (this *HallProxy) Destroy() {
	this.GameProxy.Destroy()
}

func (this *HallProxy) onUserLoginSuccess(session gate.Session,account string) {

}
