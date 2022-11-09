package common

import (
	"net/rpc"

	"github.com/lucasepe/hinge"
)

type Greeter interface {
	Greet() string
}

type GreeterRPC struct {
	client *rpc.Client
}

func (g *GreeterRPC) Greet() string {
	var resp string
	err := g.client.Call("Plugin.Greet", new(interface{}), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

type GreeterRPCServer struct {
	Impl Greeter
}

func (s *GreeterRPCServer) Greet(args interface{}, resp *string) error {
	*resp = s.Impl.Greet()
	return nil
}

type GreeterPlugin struct {
	Impl Greeter
}

func (p *GreeterPlugin) Server(*hinge.MuxBroker) (interface{}, error) {
	return &GreeterRPCServer{Impl: p.Impl}, nil
}

func (GreeterPlugin) Client(b *hinge.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &GreeterRPC{client: c}, nil
}
