package tests

import (
    "testing"
    "time"
    "enc_socks/relay"
)

func TestLocal(t *testing.T) {
    cfg := &ServerConfig{
        LocalAddr:"127.0.0.1:8847", RemoteAddr:"127.0.0.1:8848",
        Timeout:3 * time.Second, User:*relay.NewAuthMsg("xxxsen", "hello test"), ServerType:SERVER_TYPE_LOCAL}
    svr := NewRelayServer(cfg)
    svr.Start()
}

