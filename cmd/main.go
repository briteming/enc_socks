package main

import (
    "flag"
    "github.com/sirupsen/logrus"
    "time"
    "enc_socks"
)

var flagLocal = flag.String("local", "127.0.0.1:8848", "local bind addr")
var flagRemote = flag.String("remote", "127.0.0.1:8849", "remote server addr")
var flagTimeout = flag.Int("timeout", 3, "connect/read timeout")
var flagType = flag.String("type", "local", "server_type:local or remote")
var flagSvrPem = flag.String("svr_pem", "./server.pem", "pem file path")
var flagSvrKey = flag.String("svr_key", "./server.key", "key file path")

func buildConfig(config *enc_socks.ServerConfig) {
    config.LocalAddr = *flagLocal
    config.RemoteAddr = *flagRemote
    config.Timeout = time.Duration(*flagTimeout) * time.Second
    if *flagType == "local" {
        config.ServerType = enc_socks.SERVER_TYPE_LOCAL
    } else if *flagType == "remote" {
        config.ServerType = enc_socks.SERVER_TYPE_REMOTE
    }
    config.TlsServerPemAddr = *flagSvrPem
    config.TlsServerKeyAddr = *flagSvrKey
}

func main() {
    flag.Parse()
    config := &enc_socks.ServerConfig{}
    buildConfig(config)
    logrus.Infof("Parse config:%s", config.String())
    svr := enc_socks.NewRelayServer(config)
    svr.Start()
}