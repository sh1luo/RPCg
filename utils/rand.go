package utils

import (
	"math/rand"
	"sync"
	"time"
)

var (
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	mu sync.Mutex
)

func Int63n(n int64) int64 {
	mu.Lock()
	ans := r.Int63n(n)
	mu.Unlock()
	return ans
}

func Intn(n int) int {
	mu.Lock()
	ans := r.Intn(n)
	mu.Unlock()
	return ans
}