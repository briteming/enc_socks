package server

import (
    "testing"
    "time"
)

func TestServer(t *testing.T) {
    addr := "192.168.123.63:1082"
    //addr := "127.0.0.1:8849"
    cfg := &ServerConfig{"127.0.0.1:8848", addr, "hello_world", 3 * time.Second, "xor"}
    svr := NewRelayServer(cfg)
    svr.Start()

}

