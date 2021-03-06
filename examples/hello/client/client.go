package main

import (
	"context"
	"fmt"
	"rpcg/examples/hello"
	"rpcg/xclient"
)

func main() {
	d := xclient.NewRegistryDiscovery(hello.RegistryAddr, 0)
	xc := xclient.NewXClient(d, xclient.Random, nil)
	defer func() { _ = xc.Close() }()

	var reply string
	err := xc.Call(context.Background(), "HelloService.Hello", &hello.Args{Name: "shiluo"}, &reply)
	if err != nil {
		fmt.Println("xc.call err:", err)
		return
	}
	fmt.Println(reply)
}
