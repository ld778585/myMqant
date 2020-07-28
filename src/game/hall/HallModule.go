package hall

import (
	"game/define"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
)

var NewModule = func() module.Module {
	module := new(HallModule)
	return module
}

type HallModule struct {
	basemodule.BaseModule
	proxy *HallProxy
}

func (this *HallModule) Version() string {
	return define.VERSION
}

func (this *HallModule) GetType() string {
	return define.SERVER_TYPE_HALL
}

func (this *HallModule) OnInit(app module.App, settings *conf.ModuleSettings) {
	this.BaseModule.OnInit(this, app, settings)
	this.proxy = NewHallProxy(this)
}

func (this *HallModule) Run(closeSig chan bool) {
	log.Info("%v模块运行中...", this.GetType())
	this.proxy.Run()
	<-closeSig
	log.Info("%v模块已停止...", this.GetType())
}
