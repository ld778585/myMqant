package dbBSvr

import (
	"game/define"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	basemodule "github.com/liangdas/mqant/module/base"
)

var NewModule = func() module.Module {
	module := &DBSvrModule{}
	return module
}

type DBSvrModule struct {
	basemodule.BaseModule
	proxy *DBSvrProxy
}

func (this *DBSvrModule) Version() string {
	return define.VERSION
}

func (this *DBSvrModule) GetType() string {
	return define.SERVER_TYPE_DBSVR
}

func (this *DBSvrModule) OnInit(app module.App, settings *conf.ModuleSettings) {
	this.BaseModule.OnInit(this, app, settings)
	this.proxy = NewDBSvrProxy(this)
}

func (this *DBSvrModule) Run(closeSig chan bool) {
	log.Info("%v模块运行中...", this.GetType())
	this.proxy.Run()
	<-closeSig
	log.Info("%v模块已停止...", this.GetType())
}
