package gate

import (
	"fmt"
	"game/define"
	"game/msgType"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"redisClient"
	"time"
)

func NewGateProxy(module *GateModule) *GateProxy {
	proxy := new(GateProxy)
	proxy.gateModule = module
	proxy.init()
	return proxy
}

type GateProxy struct {
	gateModule *GateModule
}

func (this *GateProxy) init() {

}

func (this *GateProxy) Connect(session gate.Session) {
	log.Debug("client connect from %v", session.GetIP())
}

func (this *GateProxy) DisConnect(session gate.Session) {
	log.Debug("client disconnect,sessionId:%s,uid:%s,IP:%v",
		session.GetSessionID(),session.GetUserID(),session.GetIP())
	if session == nil {
		return
	}
	this.Storage(session)
	if session.GetUserID() != ""{
		this.Delete(session)
		this.BroadCastRpc(session, msgType.RPC_USER_LOGOUT.MessageName, session)
	}
}

func (this *GateProxy) Storage(session gate.Session) (err error) {
	data, err := session.Serializable()
	if err != nil {
		fmt.Errorf("session serializable error:%s", err.Error())
		return err
	}

	key := redisClient.GetSessionKey(session.GetUserID())
	redisClient.Expipeat(key,0)
	sOk, err := redisClient.Set(key, string(data[:]))
	if err != nil || false == sOk {
		fmt.Errorf("redis set failed:%s", err.Error())
		return err
	}
	return nil
}

func (this *GateProxy) Delete(session gate.Session) (err error) {
	key := redisClient.GetSessionKey(session.GetUserID())
	t := time.Now().Unix() + 60*10
	_, err = redisClient.Expipeat(key, t)
	if err != nil {
		log.Debug("redis del failed:", err.Error())
		return err
	}
	return nil
}

func (this *GateProxy) Query(UserID string) (data []byte, err error) {
	res, err := redisClient.Get(UserID)
	if err != nil {
		fmt.Errorf("session get from redis faild,error:%s", err.Error())
		return nil, err
	}

	return res, nil
}

func (this *GateProxy) Heartbeat(session gate.Session) {
	log.Debug("用户[%s]在线的心跳包", session.GetUserID())
}

func (this *GateProxy) OnRoute(session gate.Session, topic string, msg []byte) (bool, interface{}, error) {
	//log.Debug("onRecvMsg,topic:%s,msgLen:byte",topic,len(msg))

	res := make([]byte,64)
	session.Send(msgType.RPC_USER_LOGOUT.SendMsgName,res)
	return false, nil, nil
}

//广播到玩家所在的所有服务器
func (this *GateProxy) BroadCastRpc(session gate.Session, messageName string, args ...interface{}) {
	for _, name := range define.SERVER_NAMES {
		sid := session.GetRouteServerID(name)
		if sid == "" {
			continue
		}
		if e := this.gateModule.InvokeNR(sid, messageName, args...); e != nil {
			log.Error("broad cast rpc error,sid:%v,error:%v", sid, e)
		}
	}
}


