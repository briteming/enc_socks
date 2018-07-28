package main

import (
    "flag"
    "github.com/sirupsen/logrus"
    "time"
    "enc_socks"
    "os"
    "enc_socks/pipeline"
    "strings"
    "io/ioutil"
    "encoding/json"
)

var flagLocal = flag.String("local", "127.0.0.1:8847", "local bind addr")
var flagRemote = flag.String("remote", "127.0.0.1:8848", "remote server addr")
var flagTimeout = flag.Int("timeout", 3, "connect/read timeout")
var flagType = flag.String("type", "local", "server_type:local or remote")
var flagPipeList = flag.String("pipeline", "xor", "pipeline split with '#', example:xor#auth")
var flagPipeArgs = flag.String("pipe_args_file", "/home/sen/GoPath/src/enc_socks/cmd/config.json", "config file")

func buildConfig(config *enc_socks.ServerConfig) {
    config.LocalAddr = *flagLocal
    config.RemoteAddr = *flagRemote
    config.Timeout = time.Duration(*flagTimeout) * time.Second
    if *flagType == "local" {
        config.ServerType = enc_socks.SERVER_TYPE_LOCAL
    } else if *flagType == "remote" {
        config.ServerType = enc_socks.SERVER_TYPE_REMOTE
    }
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
    //初始化日志
    initLog()

    //生成基础配置
    config := &enc_socks.ServerConfig{}
    buildConfig(config)
    logrus.Infof("Parse config:%s", config.String())

    //加载管道配置
    mp := make(map[string]interface{})
    data, err := ioutil.ReadFile(*flagPipeArgs)
    if err != nil {
        logrus.Errorf("Load pipe args file failed, err:%s", err.Error())
    } else {
        err = json.Unmarshal(data, &mp)
        if err != nil {
            logrus.Errorf("Parse pipe file json err, msg:%s", err.Error())
            os.Exit(2)
        }
    }

    //初始化管道
    holder := pipeline.GetHolder()
    pipes := strings.Split(*flagPipeList, "#")
    var clientPipe []pipeline.Pipeline
    var serverPipe []pipeline.Pipeline
    for _, pipe := range pipes {
        fac, err := holder.GetByName(pipe)
        if err != nil {
            logrus.Errorf("Get pipe factory failed, err:%s", err.Error())
            os.Exit(1)
        }
        cfg, ok := mp[pipe]
        if !ok {
            logrus.Errorf("Get pipe:%s setting failed, setting not found!", pipe)
            os.Exit(3)
        }
        c := cfg.(map[string]interface{})
        err = fac.Init(&c)
        if err != nil {
            logrus.Errorf("Factory:%s init with setting:%v failed, err:%s", pipe, c, err.Error())
        }
        clientPipe = append(clientPipe, fac.GetCliPipe())
        serverPipe = append(serverPipe, fac.GetSvrPipe())
    }

    svr := enc_socks.NewRelayServer(config, serverPipe, clientPipe)
    svr.Start()
}