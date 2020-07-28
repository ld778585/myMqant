package main

import (
	"fmt"
	"game/gate"
	"game/login"
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/registry/consul"
	"github.com/liangdas/mqant/selector"
	"github.com/nats-io/nats.go"
	"math/rand"
	"redisClient"
	"sync"
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

	initGlobalSelector(MyApp)
	MyApp.OnConfigurationLoaded(onConfigLoaded)
	MyApp.OnStartup(onStartup)
	app.Run(
		gate.NewModule(),
		login.NewModule(),
	)
}

func onConfigLoaded(app module.App) {
	log.Debug("on configLoaded.....")
}

func onStartup(app module.App) {
	log.Debug("on startup.....")
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func initGlobalSelector(app module.App) {
	app.Options().Selector.Init(selector.SetStrategy(func(services []*registry.Service) selector.Next {
		var nodes []*registry.Node

		// Filter the nodes for datacenter
		for _, service := range services {
			for _, node := range service.Nodes {
				log.Info("server state:%v", node.Metadata["state"])
				if node.Metadata["state"] == "inline" || node.Metadata["state"] == "" {
					nodes = append(nodes, node)
				}
			}
		}

		var mtx sync.Mutex
		//log.Info("services[0] $v",services[0].Nodes[0])
		return func() (*registry.Node, error) {
			mtx.Lock()
			defer mtx.Unlock()
			if len(nodes) == 0 {
				return nil, fmt.Errorf("no node")
			}
			index := rand.Intn(int(len(nodes)))
			return nodes[index], nil
		}
	}))
}
