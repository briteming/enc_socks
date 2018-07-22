package main

import (
    "flag"
    "github.com/sirupsen/logrus"
    "time"
    "enc_socks"
    "enc_socks/relay"
    "os"
)

var flagLocal = flag.String("local", "127.0.0.1:8847", "local bind addr")
var flagRemote = flag.String("remote", "127.0.0.1:8848", "remote server addr")
var flagTimeout = flag.Int("timeout", 3, "connect/read timeout")
var flagType = flag.String("type", "local", "server_type:local or remote")
var flagSvrPem = flag.String("svr_pem", "./server.pem", "pem file path")
var flagSvrKey = flag.String("svr_key", "./server.key", "key file path")
var flagUser = flag.String("user", "xxxsen", "user name")
var flagPwd = flag.String("pwd", "hello test", "user pwd")

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
    auth := relay.NewAuthMsg(*flagUser, *flagPwd)
    config.User = *auth
    authMap := relay.NewAuthMap()
    authMap.Add(auth)
    config.UserInfo = *authMap
}

func initLog() {
    customFormatter := new(logrus.TextFormatter)
    customFormatter.FullTimestamp = true                    // 显示完整时间
    customFormatter.TimestampFormat = "2006-01-02 15:04:05" // 时间格式
    customFormatter.DisableTimestamp = false                // 禁止显示时间
    customFormatter.DisableColors = false                   // 禁止颜色显示
    logrus.SetFormatter(customFormatter)
    logrus.SetOutput(os.Stdout)
    logrus.SetLevel(logrus.DebugLevel)
}

func main() {
    flag.Parse()
    initLog()
    config := &enc_socks.ServerConfig{}
    buildConfig(config)
    logrus.Infof("Parse config:%s", config.String())
    svr := enc_socks.NewRelayServer(config)
    svr.Start()
}