package gate

import "github.com/liangdas/mqant/gate"

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

func (this *GateProxy) OnRoute(session gate.Session, topic string, msg []byte) (bool, interface{}, error) {
	panic("implement me")
}

func (this *GateProxy) Connect(a gate.Session) {
	panic("implement me")
}

func (this *GateProxy) DisConnect(a gate.Session) {
	panic("implement me")
}

func (this *GateProxy) Storage(session gate.Session) (err error) {
	panic("implement me")
}

func (this *GateProxy) Delete(session gate.Session) (err error) {
	panic("implement me")
}

func (this *GateProxy) Query(Userid string) (data []byte, err error) {
	panic("implement me")
}

func (this *GateProxy) Heartbeat(session gate.Session) {
	panic("implement me")
}


