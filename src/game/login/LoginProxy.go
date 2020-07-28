package login

import (
	"game/gameProxy"
	"game/msgType"
	"game/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	mqrpc "github.com/liangdas/mqant/rpc"
	"strconv"
)

func NewLoginProxy(module *LoginModule) *LoginProxy {
	proxy := new(LoginProxy)
	proxy.module = module
	proxy.mgr = NewLoginMgr(proxy)
	proxy.Init(proxy, &module.BaseModule, module)
	return proxy
}

type LoginProxy struct {
	gameProxy.GameProxy
	module *LoginModule
	mgr    *LoginMgr
}

func (this *LoginProxy) AddEvents() {
}

func (this *LoginProxy) RemoveEvents() {
}

func (this *LoginProxy) RegisterMessages() {
	this.RegisterGO(msgType.CS_USER_LOGIN, this.onLoginRequest)
}

func (this *LoginProxy) CancelMessages() {
}

func (this *LoginProxy) Destroy() {
	this.GameProxy.Destroy()
}

func (this *LoginProxy) onLoginRequest(session gate.Session, msg []byte) {
	if session.GetUserIDInt64() != 0{
		return
	}

	loginRequest := &protobuf.LoginRequest{}
	proto.Unmarshal(msg, loginRequest)
	log.Debug("user login,account:%v", loginRequest.Account)

	res := &protobuf.LoginResponse{}
	uid, e1 := this.mgr.onUserLogin(loginRequest.Account, loginRequest.PassWord)
	if !e1 {
		res.ErrorCode = 1
		this.SendWithProto(session, msgType.CS_USER_LOGIN, res)
		return
	}

	session.Bind(strconv.Itoa(int(uid)))
	bo, er := this.Call(session, msgType.RPC_LOAD_USER_INFO_FROM_DB, mqrpc.Param(session.GetUserIDInt64()))
	if er != "" || !bo.(bool) {
		log.Warning("Error:%v,result:%v", er, bo)
		return
	}

	e := this.InvokeNR(session, msgType.RPC_USER_LOGIN_SUCCESS, session, loginRequest.Account)
	if e != nil {
		log.Warning("Error:%v", e)
		return
	}

	this.SendCBWithProto(session, msgType.CS_USER_LOGIN, res)
}

func (this *LoginProxy) onUserLogout(session gate.Session) {

}
