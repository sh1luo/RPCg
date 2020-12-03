package main

import (
	"fmt"
	"net"
	"rpcg"
	"rpcg/examples/hello"
	"rpcg/registry"
	"time"
)

// 记录了当前服务器编号，暂且随机一个
var serverSeq int

func init() {
	serverSeq = hello.R.Int()
}

type HelloService int

type Args struct{ Name string }

func (h HelloService) Hello(args Args, reply *string) error {
	*reply = fmt.Sprintf("Server - %d reply: hello, %s!", serverSeq, args.Name)
	return nil
}

func (h HelloService) DelayHello(args Args, reply *string) error {
	time.Sleep(time.Second * 5)
	*reply = fmt.Sprintf("Server - %d reply: hello, %s!", serverSeq, args.Name)
	return nil
}

func startServer(registryAddr string) {
	var h HelloService
	l, _ := net.Listen("tcp", ":0")
	server := rpcg.NewServer()
	_ = server.Register(&h)
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), "weight=1", 0)
	server.Accept(l)
}

func main() {
	startServer(hello.RegistryAddr)
}
