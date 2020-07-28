package gameProxy

import (
	interfaces "game/interface"
	"game/msgType"
	"github.com/golang/protobuf/proto"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
	mqrpc "github.com/liangdas/mqant/rpc"
	"github.com/liangdas/mqant/selector"
	"github.com/liangdas/mqant/server"
)

type GameProxy struct {
	module   *basemodule.BaseModule
	subClass module.RPCModule
	server   server.Server
	child    interfaces.IProxy
}

func (this *GameProxy) Init(child interfaces.IProxy, module *basemodule.BaseModule, subClass module.RPCModule) {
	this.module = module
	this.subClass = subClass
	this.server = this.module.GetServer()
	this.child = child
}

func (this *GameProxy) Run() {
	this.child.AddEvents()
	this.child.RegisterMessages()
}

func (this *GameProxy) BroadCastModule(moduleType string, messageName string, args ...interface{}) {
	services := this.module.App.GetServersByType(moduleType)
	for _, server := range services {
		if server.GetID() == this.module.GetServer().Id() {
			continue
		}
		server.CallNR(messageName, args...)
	}
}

func (this *GameProxy) ResetServerID(session gate.Session, moduleType string) string {
	if session == nil {
		return moduleType
	}

	serverId := session.GetRouteServerID(moduleType)
	if serverId != "" {
		return serverId
	}

	var server, _ = this.module.App.GetRouteServer(moduleType)
	if server != nil {
		session.SetRouteServerID(moduleType, server.GetID())
		moduleType = server.GetID()
		session.Push()
	}
	return moduleType
}

func (this *GameProxy) ResetServerIdS(sessions []gate.Session, moduleType string) string {
	var server, _ = this.module.App.GetRouteServer(moduleType)
	if server != nil {
		for _, session := range sessions {
			session.SetRouteServerID(moduleType, server.GetID())
			session.Push()
		}
		moduleType = server.GetID()
	}
	return moduleType
}

//注册消息
func (this *GameProxy) Call(session gate.Session, msgType *msgType.MessageType, param mqrpc.ParamOption, opts ...selector.SelectOption) (interface{}, string) {
	moduleType := this.ResetServerID(session, msgType.ModuleName)
	return this.module.Call(nil, moduleType, msgType.MessageName, param, opts...)
}

func (this *GameProxy) Invoke(session gate.Session, msgType *msgType.MessageType, params ...interface{}) (result interface{}, err string) {
	moduleType := this.ResetServerID(session, msgType.ModuleName)
	return this.module.Invoke(moduleType, msgType.MessageName, params...)
}

func (this *GameProxy) InvokeNR(session gate.Session, msgType *msgType.MessageType, params ...interface{}) (err error) {
	moduleType := this.ResetServerID(session, msgType.ModuleName)
	return this.module.InvokeNR(moduleType, msgType.MessageName, params...)
}

func (this *GameProxy) InvokeArgs(session gate.Session, msgType *msgType.MessageType, ArgsType []string, args [][]byte) (result interface{}, err string) {
	moduleType := this.ResetServerID(session, msgType.ModuleName)
	return this.module.InvokeArgs(moduleType, msgType.MessageName, ArgsType, args)
}

func (this *GameProxy) InvokeNRArgs(session gate.Session, msgType *msgType.MessageType, ArgsType []string, args [][]byte) (err error) {
	moduleType := this.ResetServerID(session, msgType.ModuleName)
	return this.module.InvokeNRArgs(moduleType, msgType.MessageName, ArgsType, args)
}

func (this *GameProxy) RegisterGO(msgType *msgType.MessageType, f interface{}) {
	this.module.GetServer().RegisterGO(msgType.MessageName, f)
}

func (this *GameProxy) Send(session gate.Session, msgType *msgType.MessageType, body []byte) string {
	return session.Send(msgType.SendMsgName, body)
}

func (this *GameProxy) SendWithProto(session gate.Session, msgType *msgType.MessageType, body proto.Message) string {
	b, _ := proto.Marshal(body)
	return session.Send(msgType.SendMsgName, b)
}

func (this *GameProxy) SendCB(session gate.Session, msgType *msgType.MessageType, body []byte) (err string) {
	return session.SendCB(msgType.SendMsgName, body)
}
func (this *GameProxy) SendCBWithProto(session gate.Session, msgType *msgType.MessageType, body proto.Message) (err string) {
	b, _ := proto.Marshal(body)
	return session.SendCB(msgType.SendMsgName, b)
}

func (this *GameProxy) SendNR(session gate.Session, msgType *msgType.MessageType, body []byte) string {
	return session.SendNR(msgType.SendMsgName, body)
}
func (this *GameProxy) SendBatch(sessionIds string, session gate.Session, msgType *msgType.MessageType, body []byte) (int64, string) {
	return session.SendBatch(sessionIds, msgType.SendMsgName, body)
}
func (this *GameProxy) Destroy() {
	this.child.RemoveEvents()
	this.child.CancelMessages()
	this.module = nil
	this.server = nil
	this.child = nil
}