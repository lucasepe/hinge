package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/lucasepe/hinge"
	"github.com/lucasepe/hinge/example/basic/common"
)

var pluginMap = map[string]hinge.Plugin{
	"greeter": &common.GreeterPlugin{},
}

func main() {
	client := hinge.NewClient(&hinge.ClientConfig{
		Plugins: pluginMap,
		Cmd:     exec.Command("./plugin/greeter"),
	})
	defer client.Kill()

	rpcClient, err := client.Protocol()
	if err != nil {
		log.Fatal(err)
	}

	raw, err := rpcClient.Dispense("greeter")
	if err != nil {
		log.Fatal(err)
	}

	greeter := raw.(common.Greeter)
	fmt.Println(greeter.Greet())
}
