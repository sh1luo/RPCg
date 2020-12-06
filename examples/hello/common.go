package hello

import (
	"math/rand"
	"time"
)

var (
	RegistryAddr = "http://localhost:9999/_rpcg_/registry"
	R            = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type Args struct{ Name string }
