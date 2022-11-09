package main

import (
	"github.com/lucasepe/hinge"
	"github.com/lucasepe/hinge/example/basic/common"
)

type GreeterHello struct{}

func (g *GreeterHello) Greet() string {
	return "Hello!"
}

func main() {
	greeter := &GreeterHello{}

	var pluginMap = map[string]hinge.Plugin{
		"greeter": &common.GreeterPlugin{Impl: greeter},
	}
	hinge.Serve(&hinge.ServeConfig{
		Plugins: pluginMap,
	})
}
