package main

import (
	"net"
	"net/http"
	"rpcg/registry"
)

func main() {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	_ = http.Serve(l, nil)
}
