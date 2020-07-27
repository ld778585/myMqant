package gate

import (
	"game/define"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	basegate "github.com/liangdas/mqant/gate/base"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
)

var NewModule = func() module.Module {
	module := new(GateModule)
	module.init()
	return module
}

type GateModule struct {
	basegate.Gate
	proxy *GateProxy
}

func (this *GateModule) init() {
	this.proxy = NewGateProxy(this)
}

func (this *GateModule) Version() string {
	return define.VERSION
}

func (this *GateModule) GetType() string {
	return define.SERVER_TYPE_GATE
}

func (this *GateModule) OnInit(app module.App, settings *conf.ModuleSettings) {
	this.Gate.OnInit(this, app, settings,
		gate.SetStorageHandler(this.proxy),
		gate.SetSessionLearner(this.proxy),
		gate.SetRouteHandler(this.proxy),
	)
	log.Info("%v模块初始化完成...version:%v", this.GetType(), this.Version())
}

func (this *GateModule) Run(closeSig chan bool) {
	log.Info("%v模块运行中...", this.GetType())
	this.Gate.Run(closeSig)
	log.Info("%v模块已停止...", this.GetType())
}
