package server

import (
    "testing"
    "time"
)

func TestLocal(t *testing.T) {
    cfg := &ServerConfig{"127.0.0.1:8847", "127.0.0.1:8848", "hello_world", 3 * time.Second, "xor"}
    svr := NewRelayLocal(cfg)
    svr.Start()
}

