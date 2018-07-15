package main

import (
    "flag"
    "enc_socks/server"
    "github.com/sirupsen/logrus"
    "time"
    "enc_socks/codec"
    "os"
)

var flagLocal = flag.String("local", "127.0.0.1:8848", "local bind addr")
var flagRemote = flag.String("remote", "127.0.0.1:8849", "remote server addr")
var flagTimeout = flag.Int("timeout", 3, "connect/read timeout")
var flagKey = flag.String("key", "hello_test", "encrypt key")
var flagCodec = flag.String("codec", "xor", "codec")
var flagType = flag.String("type", "local", "server_type:local or remote")

func buildConfig(config *server.ServerConfig) {
    config.LocalAddr = *flagLocal
    config.RemoteAddr = *flagRemote
    config.Codec = *flagCodec
    config.Key = *flagKey
    config.Timeout = time.Duration(*flagTimeout) * time.Second
}

func main() {
    flag.Parse()
    config := &server.ServerConfig{}
    buildConfig(config)
    logrus.Infof("Parse config:%s", config.String())
    comm := codec.NewFactory()
    _, err := comm.GetFactory(config.Codec)
    if err != nil {
        logrus.Errorf("Get codec factory failed, err:%s, codec name:%s, codec list:%v", err.Error(), config.Codec, comm.AllKey())
        os.Exit(1);
    }
    if *flagType == "local" {
        svr := server.NewRelayLocal(config)
        svr.Start()
    } else {
        svr := server.NewRelayServer(config)
        svr.Start()
    }
}