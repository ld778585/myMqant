package main

import (
	"game/gate"
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/registry/consul"
	"github.com/nats-io/nats.go"
	"redisClient"
	"systemConf"
	"time"
)

var MyApp module.App

func main() {
	defer func() {
		if e := recover(); e != nil {
			log.Error("[SE] System Error:%v", e)
		}
	}()

	//加载系统配置
	e := systemConf.InitConfig()
	if e != nil {
		log.Error(e.Error())
		return
	}

	rs := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = systemConf.SystemConfMgr.ConsulConf.Address
		log.Info("consul Addrs %v", options.Addrs)
	})

	nc, err := nats.Connect(systemConf.SystemConfMgr.NatsConf.Address,
		nats.MaxReconnects(systemConf.SystemConfMgr.NatsConf.MaxReconnects))
	if err != nil {
		log.Error("nats error %v", err)
		return
	}
	log.Info("nats connect to %v", systemConf.SystemConfMgr.NatsConf.Address)

	redisClient.Initialize()
	MyApp = mqant.CreateApp(
		module.Debug(systemConf.SystemConfMgr.Debug),
		module.Nats(nc),
		module.Registry(rs),
		module.KillWaitTTL(time.Minute*1),
	)

	app := mqant.CreateApp(
		module.Debug(true),
		module.Nats(nc),
		module.Registry(rs),
		module.KillWaitTTL(time.Minute*1),
	)

	app.Run(
		gate.NewModule(),
	)
}
