package main

import (
	"context"
	"fmt"
	"rpcg/examples/hello"
	"rpcg/xclient"
)

func main() {
	d := xclient.NewGeeRegistryDiscovery(hello.RegistryAddr, 0)
	xc := xclient.NewXClient(d, xclient.Random, nil)
	defer func() { _ = xc.Close() }()
	// send request & receive response
	var reply string
	err := xc.Call(context.Background(), "HelloService.Hello", &hello.Args{Name: "shiluo"}, &reply)
	if err != nil {
		fmt.Println("xc.call err:", err)
		return
	}
}
