package main

import (
	"github.com/liangdas/mqant"
	"github.com/liangdas/mqant/module"
)

func main() {
	app := mqant.CreateApp(
		module.Debug(true),
		)

	app.Run(
		)
}
