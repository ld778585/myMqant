package login

import (
	"game/define"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
)

var NewModule = func() module.Module {
	module := new(LoginModule)
	return module
}

type LoginModule struct {
	basemodule.BaseModule
	proxy *LoginProxy
}

func (this *LoginModule) Version() string {
	return define.VERSION
}

func (this *LoginModule) GetType() string {
	return define.SERVER_TYPE_LOGIN
}

func (this *LoginModule) OnInit(app module.App, settings *conf.ModuleSettings) {
	this.BaseModule.OnInit(this, app, settings)
	this.proxy = NewLoginProxy(this)
}

func (this *LoginModule) Run(closeSig chan bool) {
	log.Info("%v模块运行中...", this.GetType())
	this.proxy.Run()
	<-closeSig
	log.Info("%v模块已停止...", this.GetType())
}
