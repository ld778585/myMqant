package gate

import (
	"fmt"
	"game/define"
	"game/msgType"
	"github.com/liangdas/mqant/gate"
	basegate "github.com/liangdas/mqant/gate/base"
	"github.com/liangdas/mqant/log"
	argsutil "github.com/liangdas/mqant/rpc/util"
	"github.com/pkg/errors"
	"redisClient"
	"strings"
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
	//log.Debug("client connect from %v,%d", session.GetIP(),unsafe.Sizeof(session))
}

func (this *GateProxy) DisConnect(session gate.Session) {
	//log.Debug("client disconnect,sessionId:%s,uid:%s,IP:%v",
	//	session.GetSessionID(),session.GetUserID(),session.GetIP())
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
	//log.Debug("用户[%s]在线的心跳包", session.GetUserID())
}

func (this *GateProxy) OnRoute(_session gate.Session, topic string, msg []byte) (bool, interface{}, error) {
	var msgid string
	topics := strings.Split(topic, "/")
	if len(topics) < 2 {
		errorstr := "Topic must be [moduleType@moduleID]/[handler]|[moduleType@moduleID]/[handler]/[msgid]"
		log.Error(errorstr)
		return true, nil, errors.New(errorstr)
	} else if len(topics) == 3 {
		msgid = topics[2]
	}
	var ArgsType []string = make([]string, 2)
	var args [][]byte = make([][]byte, 2)
	var moduleType = topics[0]
	var mt = moduleType
	var sid = ""
	if sid = _session.GetRouteServerID(moduleType); sid != "" {
		mt = sid
	}

	serverSession, err := this.gateModule.App.GetRouteServer(mt)
	if serverSession == nil || err != nil {
		serverSession,_ = this.gateModule.App.GetRouteServer(moduleType)
	}

	if serverSession == nil{
		msg := fmt.Sprintf("Service(type:%s) not found", topics[0])
		log.Error("%s", msg)
		return true, nil, errors.New(msg)
	}

	if sid != serverSession.GetID() {
		_session.SetRouteServerID(moduleType, serverSession.GetID())
	}

	ArgsType[1] = argsutil.BYTES
	args[1] = msg
	session := _session.Clone()
	session.SetTopic(topic)
	if msgid != "" {
		ArgsType[0] = basegate.RPCParamSessionType
		session.SetMsgID(msgid)
		b, err := session.Serializable()
		if err != nil {
			return false, nil, nil
		}
		args[0] = b
		serverSession.CallNRArgs(topics[1], ArgsType, args)
	} else {
		ArgsType[0] = basegate.RPCParamSessionType
		b, err := session.Serializable()
		if err != nil {
			return false, nil, nil
		}
		args[0] = b

		e := serverSession.CallNRArgs(topics[1], ArgsType, args)
		if e != nil {
			log.Warning("Gate RPC", e.Error())
		}
	}
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


