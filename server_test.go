package enc_socks

import (
    "testing"
    "time"
    "enc_socks/relay"
)

func TestServer(t *testing.T) {
    addr := "127.0.0.1:1081"
    mp := relay.NewAuthMap()
    mp.Add(relay.NewAuthMsg("xxxsen", "hello test"))
    cfg := &ServerConfig{
        LocalAddr:"127.0.0.1:8848", RemoteAddr:addr,
        Key:"hello_world", Timeout:3 * time.Second, UserInfo:*mp, ServerType:SERVER_TYPE_REMOTE,
        TlsServerPemAddr:"/home/sen/GoPath/src/enc_socks/cmd/server.pem", TlsServerKeyAddr:"/home/sen/GoPath/src/enc_socks/cmd/server.key"}
    svr := NewRelayServer(cfg)
    svr.Start()
}

